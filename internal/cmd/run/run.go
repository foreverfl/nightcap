package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"nightcap/internal/client/claude"
	"nightcap/internal/client/ghapp"
	"nightcap/internal/config"
	"nightcap/internal/events"
	"nightcap/internal/git"
	"nightcap/internal/indent"
	"nightcap/internal/notify"
	"nightcap/internal/path"
	"nightcap/internal/plan"
	"nightcap/internal/pr"
	"nightcap/internal/state"
	"nightcap/internal/test"
	"nightcap/internal/usage"
	"nightcap/internal/verdict"
)

// Run goes down the plan file args[0] PR by PR, running every open commit in a worktree of the frontmatter's repo, until nothing is left or a commit is not accepted.
// It fails when no plan file is named or the frontmatter leaves a key out, and when the config, git, claude or sh fails.
func Run(args []string, stdout io.Writer) (runErr error) {
	if len(args) == 0 {
		return errors.New("run: name a plan file")
	}
	planPath := args[0]
	planText, err := os.ReadFile(planPath)
	if err != nil {
		return err
	}
	meta, err := plan.ParseMeta(string(planText))
	if err != nil {
		return err
	}
	// a key left out reads as nil, while "protected: []" is an empty list that means none
	if meta.Repo == "" || meta.Test == "" || meta.Protected == nil || meta.Dependencies == nil {
		return errors.New("run: the plan's frontmatter needs repo, test, protected and dependencies")
	}
	configPath, err := path.ConfigFile()
	if err != nil {
		return err
	}
	loaded, err := config.Load(configPath)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if loaded.AppID != 0 && loaded.AppKey != "" {
		jwt, err := ghapp.JWT(loaded.AppID, loaded.AppKey)
		if err != nil {
			return fmt.Errorf("app: %w", err)
		}
		appName, err := ghapp.AppName(ctx, jwt)
		if err != nil {
			return fmt.Errorf("app: %w", err)
		}
		fmt.Fprintf(stdout, "app: %s\n", appName)
	}
	started := time.Now()
	var startUsage usage.Usage // the first /usage read before a commit, so a night that runs nothing never calls claude
	dataDir, err := path.DataDir(meta.Repo)
	if err != nil {
		return err
	}
	statePath := filepath.Join(dataDir, "state.json")
	cacheDir, err := path.CacheDir(meta.Repo)
	if err != nil {
		return err
	}
	worktreesDir := filepath.Join(cacheDir, "worktrees")
	var pickedPRs []plan.PR // the PRs tonight took, top to bottom
	// one file per run, under its start date: ~/.local/share/nightcap/blog-v2/runs/261108/225500.jsonl
	runPath := filepath.Join(dataDir, "runs", started.Format("060102"), started.Format("150405")+".jsonl")
	appendEvent(runPath, events.Event{Kind: "run_start", At: started})
	defer func() {
		sendNightSummary(ctx, loaded, meta.Repo, statePath, pickedPRs, started, startUsage)
		reason := ""
		if runErr != nil {
			reason = runErr.Error()
		} else if len(pickedPRs) > 0 {
			reason = stopReason(statePath, pickedPRs[len(pickedPRs)-1], started)
		}
		appendEvent(runPath, events.Event{Kind: "run_end", At: time.Now(), Reason: reason})
	}()

	// a PR whose commits all ended accepted lets the next pass take the PR below it
	for {
		picked, finished, err := runNextPR(ctx, stdout, loaded, meta, planPath, string(planText), statePath, worktreesDir, runPath, &startUsage)
		if picked != nil {
			pickedPRs = append(pickedPRs, *picked)
		}
		if err != nil || !finished {
			return err
		}
	}
}

// runNextPR takes the top PR that still has a commit to run and runs its commits until one is not accepted.
// picked is that PR, nil when there was none; finished is true when every commit of it ended accepted.
func runNextPR(ctx context.Context, stdout io.Writer, loaded config.Config, meta plan.Meta, planPath, planText, statePath, worktreesDir, runPath string, startUsage *usage.Usage) (picked *plan.PR, finished bool, err error) {
	saved, err := state.Load(statePath)
	if err != nil {
		return picked, false, err
	}
	var prsToDo []plan.PR
	hasPR := map[string]bool{}
	baseFor := map[string]string{}
	previousOpenBranch := ""
	stackOnto := map[string]string{}
	previousOpenURL := ""
	for _, planPR := range plan.Parse(planText) {
		baseFor[planPR.Slug] = "main"
		if previousOpenBranch != "" {
			baseFor[planPR.Slug] = previousOpenBranch
		}
		stackOnto[planPR.Slug] = previousOpenURL
		previousOpenBranch = ""
		previousOpenURL = ""
		if len(planPR.Commits) > 0 {
			view, err := pr.ViewBranch(meta.Repo, planPR.Type+"/"+planPR.Slug)
			if err != nil {
				return picked, false, err
			}
			if err := state.SavePRStatus(statePath, planPR.Slug, view.URL, view.State); err != nil {
				return picked, false, err
			}
			if view.State == "MERGED" || view.State == "CLOSED" {
				if err := removeWorktreeIfAny(stdout, meta.Repo, filepath.Join(worktreesDir, planPR.Slug), planPR.Slug, view.State); err != nil {
					return picked, false, err
				}
				fmt.Fprintf(stdout, "skip %s: PR %s\n", planPR.Title, strings.ToLower(view.State))
				continue
			}
			hasPR[planPR.Slug] = view.State == "OPEN"
			if hasPR[planPR.Slug] {
				previousOpenBranch = planPR.Type + "/" + planPR.Slug
				previousOpenURL = view.URL
			}
		}
		if isPRDone(saved.PRs[planPR.Slug], planPR) {
			continue // its branch above still becomes the base of the PR below
		}
		prsToDo = append(prsToDo, planPR)
	}
	planPR, ok := plan.Pick(prsToDo)
	if !ok {
		fmt.Fprintf(stdout, "%s: nothing to take\n", planPath)
		return nil, false, nil
	}
	picked = &planPR
	base := baseFor[planPR.Slug]
	branch := planPR.Type + "/" + planPR.Slug
	worktreeDir, err := git.AddWorktree(meta.Repo, base, branch, filepath.Join(worktreesDir, planPR.Slug))
	if err != nil {
		return picked, false, err
	}
	fmt.Fprintln(stdout, planPath)
	indent.Fprintf(stdout, 1, "branch: %s\n", branch)
	indent.Fprintf(stdout, 1, "worktree: %s\n", worktreeDir)
	for _, commit := range planPR.Commits {
		mark := "skip"
		if commit.HasTest {
			mark = "take"
		}
		indent.Fprintf(stdout, 1, "%s  %s\n", mark, commit.Title)
	}

	for index, commit := range planPR.Commits {
		if !commit.HasTest {
			continue
		}
		if done, _ := isAccepted(statePath, planPR.Slug, commit.Title); done {
			continue // done in an earlier run; the plan still shows it as [ ]
		}
		// a failed /usage reads as unknown usage; the percent limits are skipped and the rate-limit error is trusted instead
		usageJSON, _ := claude.Usage(ctx)
		current, _ := usage.Read(usageJSON)
		if !startUsage.Known {
			*startUsage = current
		}
		if reason := usageStopReason(time.Now(), current, loaded); reason != "" {
			indent.Fprintf(stdout, 1, "stop: %s\n", reason)
			return picked, false, state.SaveCommit(statePath, planPR.Slug, commit.Title, state.Commit{Result: "stopped", Reason: reason})
		}
		if err := runCommit(ctx, stdout, loaded, meta, worktreeDir, base, planPR.Slug, index+1, commit, statePath, runPath); err != nil {
			return picked, false, err
		}
		accepted, err := isAccepted(statePath, planPR.Slug, commit.Title)
		if err != nil {
			return picked, false, err
		}
		if !accepted {
			return picked, false, nil // red, blocked or stopped: the commits below stand on this one
		}
		if err := openPR(ctx, stdout, statePath, runPath, worktreeDir, base, stackOnto[planPR.Slug], planPR, commit.Title, hasPR[planPR.Slug], slices.Contains(loaded.Lang, "en")); err != nil {
			return picked, false, err
		}
		hasPR[planPR.Slug] = true
		if err := labelWhenFinished(stdout, statePath, worktreeDir, loaded.MergeMethod, planPR); err != nil {
			return picked, false, err
		}
	}
	return picked, true, nil
}

// isAccepted reports whether commitTitle ended green or read-code in the state file.
func isAccepted(statePath, slug, commitTitle string) (bool, error) {
	saved, err := state.Load(statePath)
	if err != nil {
		return false, err
	}
	result := saved.PRs[slug].Commits[commitTitle].Result
	return result == "green" || result == "read-code", nil
}

// isPRDone reports whether every commit of planPR that has a check ended green or read-code.
func isPRDone(savedPR state.PR, planPR plan.PR) bool {
	for _, commit := range planPR.Commits {
		result := savedPR.Commits[commit.Title].Result
		if commit.HasTest && result != "green" && result != "read-code" {
			return false
		}
	}
	return true
}

// runCommit runs claude and the test on one commit until the test is green or MaxRounds is spent.
// It prints "blocked" when the rounds run out, the same failure comes back twice, or claude stops before Done.
func runCommit(ctx context.Context, stdout io.Writer, loaded config.Config, meta plan.Meta, worktreeDir, base, slug string, number int, commit plan.Commit, statePath, runPath string) error {
	failureTail := ""
	sessionID := ""
	record := state.Commit{}
	appendEvent(runPath, events.Event{Kind: "commit_start", At: time.Now(), PR: slug, Commit: commit.Title})
	for round := 1; round <= loaded.MaxRounds; round++ {
		record.Attempts = round
		logPath := filepath.Join(filepath.Dir(statePath), "logs", slug, fmt.Sprintf("commit%d-round%d.log", number, round))
		roundStarted := time.Now()
		result, err := claude.RunRound(ctx, worktreeDir, BuildPrompt(commit.Body, failureTail, sessionID != ""), sessionID, loaded.MaxTurns, loaded.AllowedTools, loaded.AuthorName, loaded.AuthorEmail, logPath, stdout)
		if err != nil {
			return err
		}
		if result.SessionID != "" {
			sessionID = result.SessionID
		}
		appendEvent(runPath, events.Event{Kind: "round", At: time.Now(), PR: slug, Commit: commit.Title,
			Round: round, Seconds: int(time.Since(roundStarted).Seconds()), SessionID: result.SessionID})
		indent.Fprintf(stdout, 1, "ran  %s: %s (log: %s)\n", commit.Title, result.Outcome, logPath)
		if result.Outcome == claude.RateLimited {
			usageJSON, _ := claude.Usage(ctx)
			current, _ := usage.Read(usageJSON)
			if current.Known {
				record.UsageAtLimit = &state.UsageAtLimit{Session: current.Session.Percent, WeeklyAll: current.WeeklyAll.Percent}
			}
			reset, afterWork := windowReset(time.Now(), current, loaded.WorkStart)
			// the reset comes after work_start, so sleeping would carry the run into the workday
			if afterWork != "" {
				indent.Fprintf(stdout, 1, "stop: rate limited, %s\n", afterWork)
				record.Result, record.Reason = "stopped", "rate limited, "+afterWork
				return state.SaveCommit(statePath, slug, commit.Title, record)
			}
			indent.Fprintf(stdout, 1, "sleep until %s\n", reset.Local().Format("15:04"))
			sleepUntil(ctx, reset)
			continue
		}
		if result.Outcome != claude.Done {
			break
		}
		// 1. Test: run the test command from the plan's frontmatter (pass or not)
		tested, err := test.Run(ctx, worktreeDir, meta.Test)
		if err != nil {
			return err
		}
		indent.Fprintf(stdout, 1, "test: %s\n", color(tested.Green))
		appendEvent(runPath, events.Event{Kind: "test", At: time.Now(), PR: slug, Commit: commit.Title, Result: color(tested.Green)})
		failure := tested.Tail
		record.Result, record.Reason, record.TestTail = "red", "", tested.Tail
		// 2. Verdict: sort what the branch changed by the rules in verdict/classify.go
		if tested.Green {
			verdict, err := classifyBranch(stdout, worktreeDir, base, meta, commit)
			if err != nil {
				return err
			}
			record.Reason = verdict.Reason
			if !verdict.Red {
				record.Result = "green"
				if verdict.ReadCode {
					record.Result = "read-code"
				}
			}
			appendEvent(runPath, events.Event{Kind: "verdict", At: time.Now(), PR: slug, Commit: commit.Title, Result: record.Result, Reason: verdict.Reason})
			if !verdict.Red {
				return state.SaveCommit(statePath, slug, commit.Title, record)
			}
			failure = verdict.Reason
		}
		if err := state.SaveCommit(statePath, slug, commit.Title, record); err != nil {
			return err
		}
		sameFailureAsLastRound := round > 1 && failure == failureTail
		if sameFailureAsLastRound {
			break
		}
		failureTail = failure
	}
	indent.Fprintf(stdout, 1, "blocked\n")
	appendEvent(runPath, events.Event{Kind: "blocked", At: time.Now(), PR: slug, Commit: commit.Title})
	record.Result = "blocked"
	return state.SaveCommit(statePath, slug, commit.Title, record)
}

// usageStopReason says why no commit should start now under loaded's limits, or "" when one may.
func usageStopReason(now time.Time, current usage.Usage, loaded config.Config) string {
	// stop 1: the 5h window is already at limit_session, so this commit could hit the limit halfway
	if current.Known && loaded.LimitSession > 0 && current.Session.Percent >= loaded.LimitSession {
		return fmt.Sprintf("session %d%% ≥ %d", current.Session.Percent, loaded.LimitSession)
	}
	// stop 2: the week is at limit_weekly_all and its reset is over 24h away, so tonight would eat the weekdays' share
	weeklyResetsSoon := current.WeeklyAll.ResetsAt.Sub(now) < 24*time.Hour
	if current.Known && loaded.LimitWeeklyAll > 0 && !weeklyResetsSoon && current.WeeklyAll.Percent >= loaded.LimitWeeklyAll {
		return fmt.Sprintf("weekly_all %d%% ≥ %d", current.WeeklyAll.Percent, loaded.LimitWeeklyAll)
	}
	// stop 3: the window would still be running at work_start, so the morning would start rate-limited
	_, afterWork := windowReset(now, current, loaded.WorkStart)
	return afterWork
}

// windowReset returns when the 5h window resets, and afterWork such as "reset 10:00 > work_start 09:00" when that is after work_start.
func windowReset(now time.Time, current usage.Usage, workStartHHMM string) (reset time.Time, afterWork string) {
	reset = now.Add(5 * time.Hour)
	if current.Session.ResetsAt.After(now) {
		reset = current.Session.ResetsAt
	}
	workStart, _ := usage.NextWorkStart(now, workStartHHMM)
	if usage.EndsBeforeWork(reset, workStart) {
		return reset, ""
	}
	return reset, fmt.Sprintf("reset %s > work_start %s", reset.In(now.Location()).Format("15:04"), workStart.Format("15:04"))
}

// sleepUntil waits until wakeAt or until ctx is cancelled; tests swap it out so they do not wait.
var sleepUntil = func(ctx context.Context, wakeAt time.Time) {
	select {
	case <-time.After(time.Until(wakeAt)):
	case <-ctx.Done():
	}
}

// color names a test result the way the plan file does: "green" for pass, "red" for fail.
func color(green bool) string {
	if green {
		return "green"
	}
	return "red"
}

// classifyBranch prints the files the branch changed and returns the verdict on them.
func classifyBranch(stdout io.Writer, worktreeDir, base string, meta plan.Meta, commit plan.Commit) (verdict.Result, error) {
	changes, err := git.ChangedFiles(worktreeDir, base)
	if err != nil {
		return verdict.Result{}, err
	}
	for _, change := range changes {
		indent.Fprintf(stdout, 1, "%s %s\n", change.Status, change.Path)
	}
	addedLines, err := git.AddedLines(worktreeDir, base)
	if err != nil {
		return verdict.Result{}, err
	}
	indent.Fprintf(stdout, 1, "added lines: %d\n", len(addedLines))
	result := verdict.Classify(changes, addedLines, verdict.WherePaths(commit.Body), meta.Protected, meta.Dependencies)
	if result.Red {
		indent.Fprintf(stdout, 1, "verdict: red (%s)\n", result.Reason)
	} else if result.ReadCode {
		indent.Fprintf(stdout, 1, "verdict: green, read the code (%s)\n", result.Reason)
	} else {
		indent.Fprintf(stdout, 1, "verdict: green\n")
	}
	return result, nil
}

// openPR pushes planPR's branch and opens its PR, or rewrites the body when hasPR. The caller has checked that commitTitle was accepted.
func openPR(ctx context.Context, stdout io.Writer, statePath, runPath, worktreeDir, base, previousURL string, planPR plan.PR, commitTitle string, hasPR, isEnglishWanted bool) error {
	saved, err := state.Load(statePath)
	if err != nil {
		return err
	}
	branch := planPR.Type + "/" + planPR.Slug
	if err := git.Push(worktreeDir, branch); err != nil {
		return err
	}
	changes, err := git.ChangedFiles(worktreeDir, base)
	if err != nil {
		return err
	}
	addedLines, err := git.AddedLines(worktreeDir, base)
	if err != nil {
		return err
	}
	input := bodyInput(planPR, saved.PRs[planPR.Slug], commitTitle)
	input.Risks = pr.RiskLines(changes, addedLines)
	diff, err := git.Diff(worktreeDir, base)
	if err != nil {
		return err
	}
	reviewOrder, reviewOrderErr := claude.ReviewOrder(ctx, diff)
	input.ReviewOrder, input.IsMissingReviewOrder = reviewOrder, reviewOrderErr != nil
	var body string
	if isEnglishWanted {
		translate := func(korean string) (string, error) { return claude.Translate(ctx, korean) }
		body, err = pr.BuildBilingualBody(input, translate)
	} else {
		body, err = pr.BuildBody(input)
	}
	if err != nil {
		return err
	}
	if hasPR {
		return pr.EditBody(worktreeDir, branch, body)
	}
	prURL, err := pr.Create(worktreeDir, base, planPR.Type+": "+planPR.Title, body)
	if err != nil {
		return err
	}
	indent.Fprintf(stdout, 1, "PR: %s\n", prURL)
	appendEvent(runPath, events.Event{Kind: "pr_open", At: time.Now(), PR: planPR.Slug, Commit: commitTitle, URL: prURL})
	if previousURL != "" {
		if err := pr.LinkStack(worktreeDir, previousURL, prURL); err != nil {
			return err
		}
	}
	return state.SavePRStatus(statePath, planPR.Slug, prURL, "OPEN")
}

// bodyInput returns the PR body's values: planPR's summary and every commit that ended green or read-code.
func bodyInput(planPR plan.PR, savedPR state.PR, lastCommitTitle string) pr.BodyInput {
	input := pr.BodyInput{Summary: planPR.Summary, TestTail: savedPR.Commits[lastCommitTitle].TestTail}
	for _, commit := range planPR.Commits {
		record := savedPR.Commits[commit.Title]
		if record.Result != "green" && record.Result != "read-code" {
			continue
		}
		bodyCommit := pr.BodyCommit{What: plan.LabelValue(commit.Body, "WHAT"), Done: plan.LabelValue(commit.Body, "DONE")}
		if record.Result == "read-code" {
			bodyCommit.ReadCodeReason = record.Reason
		}
		input.Commits = append(input.Commits, bodyCommit)
	}
	return input
}

// labelWhenFinished adds, once every commit is done, the label the PR's mark names, and turns on auto-merge for 🤖.
func labelWhenFinished(stdout io.Writer, statePath, worktreeDir, mergeMethod string, planPR plan.PR) error {
	saved, err := state.Load(statePath)
	if err != nil {
		return err
	}
	for _, commit := range planPR.Commits {
		result := saved.PRs[planPR.Slug].Commits[commit.Title].Result
		if result != "green" && result != "read-code" {
			return nil
		}
	}
	branch := planPR.Type + "/" + planPR.Slug
	label := pr.ReadingLabel(planPR.Mark)
	if err := pr.AddLabel(worktreeDir, branch, label); err != nil {
		return err
	}
	indent.Fprintf(stdout, 1, "label: %s\n", label)
	if planPR.Mark == plan.MarkAuto {
		if err := pr.EnableAutoMerge(worktreeDir, branch, mergeMethod); err != nil {
			return err
		}
		indent.Fprintf(stdout, 1, "auto-merge: %s\n", branch)
	}
	return nil
}

// sendNightSummary sends the morning message; a failed send only prints to stderr, so it never changes the run's result.
func sendNightSummary(ctx context.Context, loaded config.Config, repo, statePath string, pickedPRs []plan.PR, started time.Time, startUsage usage.Usage) {
	summary := notify.Summary{Repos: []string{filepath.Base(repo)}, Start: started, End: time.Now()}
	for _, pickedPR := range pickedPRs {
		prResult, err := prSummary(statePath, pickedPR)
		if err != nil {
			fmt.Fprintln(os.Stderr, "notify:", err)
			continue
		}
		summary.PRs = append(summary.PRs, prResult)
	}
	if len(pickedPRs) > 0 {
		summary.StopReason = stopReason(statePath, pickedPRs[len(pickedPRs)-1], started)
	}
	if startUsage.Known {
		usageJSON, _ := claude.Usage(ctx)
		if endUsage, _ := usage.Read(usageJSON); endUsage.Known {
			summary.Usage = []notify.UsageLine{
				{Name: "session", Start: startUsage.Session.Percent, End: endUsage.Session.Percent},
				{Name: "weekly", Start: startUsage.WeeklyAll.Percent, End: endUsage.WeeklyAll.Percent},
			}
		}
	}
	if err := notify.Send(loaded.Webhook, notify.Message(summary)); err != nil {
		fmt.Fprintln(os.Stderr, "notify:", err)
	}
}

// prSummary returns planPR's line for the morning message, counting green and read-code commits as done.
func prSummary(statePath string, planPR plan.PR) (notify.PR, error) {
	saved, err := state.Load(statePath)
	if err != nil {
		return notify.PR{}, err
	}
	savedPR := saved.PRs[planPR.Slug]
	prResult := notify.PR{
		Mark:  planPR.Mark,
		Title: fmt.Sprintf("%s: %s (%s)", planPR.Type, planPR.Title, planPR.Slug),
		Total: len(planPR.Commits),
		URL:   savedPR.PRURL,
	}
	for _, commit := range planPR.Commits {
		result := savedPR.Commits[commit.Title].Result
		if result == "green" || result == "read-code" {
			prResult.Done++
		} else if result != "" && prResult.Failure == "" {
			prResult.Failure = result
		}
	}
	return prResult, nil
}

// stopReason returns why a commit of planPR stopped after started, or "" when none did during this run.
func stopReason(statePath string, planPR plan.PR, started time.Time) string {
	saved, _ := state.Load(statePath)
	for _, commit := range saved.PRs[planPR.Slug].Commits {
		if commit.Result == "stopped" && commit.UpdatedAt.After(started) {
			return commit.Reason
		}
	}
	return ""
}

// appendEvent adds event to the run file; a failed write only prints to stderr, so the log never stops a run.
func appendEvent(path string, event events.Event) {
	if err := events.Append(path, event); err != nil {
		fmt.Fprintln(os.Stderr, "events:", err)
	}
}

// removeWorktreeIfAny deletes the PR's worktree once gh says the PR is MERGED or CLOSED, and prints one line when there was one.
func removeWorktreeIfAny(stdout io.Writer, repo, worktreeDir, slug, prState string) error {
	if _, err := os.Stat(worktreeDir); err != nil {
		return nil
	}
	if err := git.RemoveWorktree(repo, worktreeDir); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "removed worktree %s: PR %s\n", slug, strings.ToLower(prState))
	return nil
}
