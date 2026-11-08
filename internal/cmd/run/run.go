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

	"nightcap/internal/check"
	"nightcap/internal/client/claude"
	"nightcap/internal/config"
	"nightcap/internal/events"
	"nightcap/internal/git"
	"nightcap/internal/notify"
	"nightcap/internal/plan"
	"nightcap/internal/pr"
	"nightcap/internal/state"
	"nightcap/internal/usage"
)

// Run is the "run" subcommand; args is everything after "run", and its first item is the plan file.
// For now it picks the plan's top heading that has a unit to take, checks its branch out in a worktree of the frontmatter's repo and runs
// the first unit that has a check through runUnit; it returns an error when no plan file is named,
// or when the config, the plan file, git, claude or sh fails.
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
	loaded, err := config.Load(".")
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	started := time.Now()
	var startUsage usage.Usage // the first /usage read before a unit, so a night that runs nothing never calls claude
	statePath := filepath.Join(meta.Repo, ".nightcap", "state.json")
	var pickedSection *plan.PR // the heading tonight took; nil when there was none
	// one file per run, under its start date: .nightcap/runs/261108/225500.jsonl
	runPath := filepath.Join(meta.Repo, ".nightcap", "runs", started.Format("060102"), started.Format("150405")+".jsonl")
	appendEvent(runPath, events.Event{Kind: "run_start", At: started})
	defer func() {
		sendNightSummary(ctx, loaded, meta.Repo, statePath, pickedSection, started, startUsage)
		reason := ""
		if runErr != nil {
			reason = runErr.Error()
		} else if pickedSection != nil {
			reason = stopReason(statePath, *pickedSection, started)
		}
		appendEvent(runPath, events.Event{Kind: "run_end", At: time.Now(), Reason: reason})
	}()

	saved, err := state.Load(statePath)
	if err != nil {
		return err
	}
	var notBlocked []plan.PR
	hasPR := map[string]bool{}
	baseFor := map[string]string{}
	previousOpenBranch := ""
	stackOnto := map[string]string{}
	previousOpenURL := ""
	for _, section := range plan.Parse(string(planText)) {
		baseFor[section.Slug] = "main"
		if previousOpenBranch != "" {
			baseFor[section.Slug] = previousOpenBranch
		}
		stackOnto[section.Slug] = previousOpenURL
		previousOpenBranch = ""
		previousOpenURL = ""
		if state.IsBlocked(saved.Sections[section.Slug]) {
			fmt.Fprintf(stdout, "stop %s: blocked, the headings below wait\n", section.Title)
			break // plans are written top to bottom, so a heading below stands on this one
		}
		if len(section.Commits) > 0 {
			view, err := pr.ViewBranch(meta.Repo, section.Type+"/"+section.Slug)
			if err != nil {
				return err
			}
			if err := state.SavePRStatus(statePath, section.Slug, view.URL, view.State); err != nil {
				return err
			}
			if view.State == "MERGED" || view.State == "CLOSED" {
				fmt.Fprintf(stdout, "skip %s: PR %s\n", section.Title, strings.ToLower(view.State))
				continue
			}
			hasPR[section.Slug] = view.State == "OPEN"
			if hasPR[section.Slug] {
				previousOpenBranch = section.Type + "/" + section.Slug
				previousOpenURL = view.URL
			}
		}
		notBlocked = append(notBlocked, section)
	}
	section, ok := plan.Pick(notBlocked)
	if !ok {
		fmt.Fprintf(stdout, "%s: nothing to take\n", planPath)
		return nil
	}
	pickedSection = &section
	base := baseFor[section.Slug]
	branch := section.Type + "/" + section.Slug
	worktreeDir, err := git.AddWorktree(meta.Repo, base, branch, section.Slug)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%s\n  branch: %s\n  worktree: %s\n", planPath, branch, worktreeDir)
	for _, unit := range section.Commits {
		mark := "skip"
		if unit.HasCheck {
			mark = "take"
		}
		fmt.Fprintf(stdout, "  %s  %s\n", mark, unit.Title)
	}

	for index, unit := range section.Commits {
		if !unit.HasCheck {
			continue
		}
		// a failed /usage reads as unknown usage; the percent limits are skipped and the rate-limit error is trusted instead
		usageJSON, _ := claude.Usage(ctx)
		current, _ := usage.Read(usageJSON)
		if !startUsage.Known {
			startUsage = current
		}
		if reason := usageStopReason(time.Now(), current, loaded); reason != "" {
			fmt.Fprintf(stdout, "  stop: %s\n", reason)
			return state.SaveUnit(statePath, section.Slug, unit.Title, state.Unit{Result: "stopped", Reason: reason})
		}
		if err := runUnit(ctx, stdout, loaded, meta, worktreeDir, base, section.Slug, index+1, unit, statePath, runPath); err != nil {
			return err
		}
		if err := openPR(ctx, stdout, statePath, runPath, worktreeDir, base, stackOnto[section.Slug], section, unit.Title, hasPR[section.Slug], slices.Contains(loaded.Lang, "en")); err != nil {
			return err
		}
		if err := labelWhenFinished(stdout, statePath, worktreeDir, loaded.MergeMethod, section); err != nil {
			return err
		}
		break
	}
	return nil
}

// runUnit runs claude then the check on one unit up to MaxRounds times, feeding a red check's tail or a red verdict's
// reason into the next prompt. It prints "blocked" when the rounds run out, the same failure comes back twice in a row,
// or claude stops before Done. Every round's outcome goes into the state file at statePath,
// and commit_start, each round and each test go as lines into the run file at runPath.
// It returns an error only when claude, sh, git or the state file fails.
func runUnit(ctx context.Context, stdout io.Writer, loaded config.Config, meta plan.Meta, worktreeDir, base, slug string, number int, unit plan.Commit, statePath, runPath string) error {
	failureTail := ""
	sessionID := ""
	record := state.Unit{}
	appendEvent(runPath, events.Event{Kind: "commit_start", At: time.Now(), PR: slug, Commit: unit.Title})
	for round := 1; round <= loaded.MaxRounds; round++ {
		record.Attempts = round
		logPath := filepath.Join(meta.Repo, ".nightcap", "logs", slug, fmt.Sprintf("unit%d-round%d.log", number, round))
		roundStarted := time.Now()
		result, err := claude.RunRound(ctx, worktreeDir, BuildPrompt(unit.Body, failureTail, sessionID != ""), sessionID, loaded.MaxTurns, loaded.AllowedTools, logPath)
		if err != nil {
			return err
		}
		if result.SessionID != "" {
			sessionID = result.SessionID
		}
		appendEvent(runPath, events.Event{Kind: "round", At: time.Now(), PR: slug, Commit: unit.Title,
			Round: round, Seconds: int(time.Since(roundStarted).Seconds()), SessionID: result.SessionID})
		fmt.Fprintf(stdout, "  ran  %s: %s (log: %s)\n", unit.Title, result.Outcome, logPath)
		if result.Outcome == claude.RateLimited {
			usageJSON, _ := claude.Usage(ctx)
			current, _ := usage.Read(usageJSON)
			if current.Known {
				record.UsageAtLimit = &state.UsageAtLimit{Session: current.Session.Percent, WeeklyAll: current.WeeklyAll.Percent}
			}
			reset, afterWork := windowReset(time.Now(), current, loaded.WorkStart)
			// the reset comes after work_start, so sleeping would carry the run into the workday
			if afterWork != "" {
				fmt.Fprintf(stdout, "  stop: rate limited, %s\n", afterWork)
				record.Result, record.Reason = "stopped", "rate limited, "+afterWork
				return state.SaveUnit(statePath, slug, unit.Title, record)
			}
			fmt.Fprintf(stdout, "  sleep until %s\n", reset.Local().Format("15:04"))
			sleepUntil(ctx, reset)
			continue
		}
		if result.Outcome != claude.Done {
			break
		}
		// 1. Check: run the check command from nightcap.yml (tests pass or not)
		checked, err := check.Run(ctx, worktreeDir, meta.Test)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "  check: %s\n", color(checked.Green))
		appendEvent(runPath, events.Event{Kind: "test", At: time.Now(), PR: slug, Commit: unit.Title, Result: color(checked.Green)})
		failure := checked.Tail
		record.Result, record.Reason, record.CheckTail = "red", "", checked.Tail
		// 2. Classify: sort what the branch changed by the rules in check/classify.go
		if checked.Green {
			verdict, err := classifyBranch(stdout, worktreeDir, base, meta, unit)
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
			appendEvent(runPath, events.Event{Kind: "verdict", At: time.Now(), PR: slug, Commit: unit.Title, Result: record.Result, Reason: verdict.Reason})
			if !verdict.Red {
				return state.SaveUnit(statePath, slug, unit.Title, record)
			}
			failure = verdict.Reason
		}
		if err := state.SaveUnit(statePath, slug, unit.Title, record); err != nil {
			return err
		}
		sameFailureAsLastRound := round > 1 && failure == failureTail
		if sameFailureAsLastRound {
			break
		}
		failureTail = failure
	}
	fmt.Fprintln(stdout, "  blocked")
	appendEvent(runPath, events.Event{Kind: "blocked", At: time.Now(), PR: slug, Commit: unit.Title})
	record.Result = "blocked"
	return state.SaveUnit(statePath, slug, unit.Title, record)
}

// usageStopReason says why no unit should open at now under loaded's limits, or "" when one may.
// The weekly limit is skipped within 24h of its reset, and a window not open yet is taken to reset at now + 5h.
func usageStopReason(now time.Time, current usage.Usage, loaded config.Config) string {
	// stop 1: the 5h window is already at limit_session, so this unit could hit the limit halfway
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

// windowReset returns when the 5h window resets (the time /usage reported, or now + 5h when no window is open or /usage failed),
// and afterWork "reset 10:00 > work_start 09:00" when that is after work_start, or "".
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

// color names a check result the way the plan file does: "green" for pass, "red" for fail.
func color(green bool) string {
	if green {
		return "green"
	}
	return "red"
}

// classifyBranch prints the files and the count of lines worktreeDir's branch changed, then the verdict on them.
// It returns that verdict, or an error when git fails.
func classifyBranch(stdout io.Writer, worktreeDir, base string, meta plan.Meta, unit plan.Commit) (check.Verdict, error) {
	changes, err := git.ChangedFiles(worktreeDir, base)
	if err != nil {
		return check.Verdict{}, err
	}
	for _, change := range changes {
		fmt.Fprintf(stdout, "  %s %s\n", change.Status, change.Path)
	}
	addedLines, err := git.AddedLines(worktreeDir, base)
	if err != nil {
		return check.Verdict{}, err
	}
	fmt.Fprintf(stdout, "  added lines: %d\n", len(addedLines))
	verdict := check.Classify(changes, addedLines, check.WherePaths(unit.Body), meta.Protected, meta.Dependencies)
	if verdict.Red {
		fmt.Fprintf(stdout, "  verdict: red (%s)\n", verdict.Reason)
	} else if verdict.ReadCode {
		fmt.Fprintf(stdout, "  verdict: green, read the code (%s)\n", verdict.Reason)
	} else {
		fmt.Fprintln(stdout, "  verdict: green")
	}
	return verdict, nil
}

// openPR pushes section's branch from worktreeDir once unitTitle has ended green or read-code in the state file at statePath,
// then opens a PR, or with hasPR replaces its description, listing every finished unit (English above Korean when
// isEnglishWanted). A new PR on top of the open PR at previousURL joins its stack, and a pr_open line goes into the run file at runPath.
// It returns an error when git, gh or the state file fails.
func openPR(ctx context.Context, stdout io.Writer, statePath, runPath, worktreeDir, base, previousURL string, section plan.PR, unitTitle string, hasPR, isEnglishWanted bool) error {
	saved, err := state.Load(statePath)
	if err != nil {
		return err
	}
	result := saved.Sections[section.Slug].Units[unitTitle].Result
	isAccepted := result == "green" || result == "read-code"
	if !isAccepted {
		return nil
	}
	branch := section.Type + "/" + section.Slug
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
	input := bodyInput(section, saved.Sections[section.Slug], unitTitle)
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
	prURL, err := pr.Create(worktreeDir, base, section.Type+": "+section.Title, body)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "  PR: %s\n", prURL)
	appendEvent(runPath, events.Event{Kind: "pr_open", At: time.Now(), PR: section.Slug, Commit: unitTitle, URL: prURL})
	if previousURL != "" {
		if err := pr.LinkStack(worktreeDir, previousURL, prURL); err != nil {
			return err
		}
	}
	return state.SavePRStatus(statePath, section.Slug, prURL, "OPEN")
}

// bodyInput collects the PR description's values: section's summary, every unit of section that ended green or read-code
// in sectionState, in plan order, and the last check tail of lastUnitTitle.
func bodyInput(section plan.PR, sectionState state.Section, lastUnitTitle string) pr.BodyInput {
	input := pr.BodyInput{Summary: section.Summary, CheckTail: sectionState.Units[lastUnitTitle].CheckTail}
	for _, unit := range section.Commits {
		record := sectionState.Units[unit.Title]
		if record.Result != "green" && record.Result != "read-code" {
			continue
		}
		bodyUnit := pr.BodyUnit{What: plan.LabelValue(unit.Body, "WHAT"), Done: plan.LabelValue(unit.Body, "DONE")}
		if record.Result == "read-code" {
			bodyUnit.ReadCodeReason = record.Reason
		}
		input.Units = append(input.Units, bodyUnit)
	}
	return input
}

// labelWhenFinished adds, once every unit is done, the label the heading's mark names, and turns on auto-merge for 🤖.
func labelWhenFinished(stdout io.Writer, statePath, worktreeDir, mergeMethod string, section plan.PR) error {
	saved, err := state.Load(statePath)
	if err != nil {
		return err
	}
	for _, unit := range section.Commits {
		result := saved.Sections[section.Slug].Units[unit.Title].Result
		if result != "green" && result != "read-code" {
			return nil
		}
	}
	branch := section.Type + "/" + section.Slug
	label := pr.ReadingLabel(section.Mark)
	if err := pr.AddLabel(worktreeDir, branch, label); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "  label: %s\n", label)
	if section.Mark == plan.MarkAuto {
		if err := pr.EnableAutoMerge(worktreeDir, branch, mergeMethod); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "  auto-merge: %s\n", branch)
	}
	return nil
}

// sendNightSummary sends the morning message: the repo, the times, when a unit ran the usage, and the picked heading's line
// read from the state file at statePath. A failed send or an unreadable state file only prints to stderr.
func sendNightSummary(ctx context.Context, loaded config.Config, repo, statePath string, pickedSection *plan.PR, started time.Time, startUsage usage.Usage) {
	summary := notify.Summary{Repos: []string{filepath.Base(repo)}, Start: started, End: time.Now()}
	if pickedSection != nil {
		prResult, err := prSummary(statePath, *pickedSection)
		if err != nil {
			fmt.Fprintln(os.Stderr, "notify:", err)
		} else {
			summary.PRs = []notify.PR{prResult}
		}
		summary.StopReason = stopReason(statePath, *pickedSection, started)
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

// prSummary reads how section's units ended in the state file at statePath: green and read-code count as done,
// and the first other result is the failure word. It returns an error when the state file cannot be read.
func prSummary(statePath string, section plan.PR) (notify.PR, error) {
	saved, err := state.Load(statePath)
	if err != nil {
		return notify.PR{}, err
	}
	sectionState := saved.Sections[section.Slug]
	prResult := notify.PR{
		Mark:  section.Mark,
		Title: fmt.Sprintf("%s: %s (%s)", section.Type, section.Title, section.Slug),
		Total: len(section.Commits),
		URL:   sectionState.PRURL,
	}
	for _, unit := range section.Commits {
		result := sectionState.Units[unit.Title].Result
		if result == "green" || result == "read-code" {
			prResult.Done++
		} else if result != "" && prResult.Failure == "" {
			prResult.Failure = result
		}
	}
	return prResult, nil
}

// stopReason returns the reason of a unit of section that the state file at statePath saved as stopped after started,
// so a stop from an earlier night is not reported again; it returns "" when there is none or the file cannot be read.
func stopReason(statePath string, section plan.PR, started time.Time) string {
	saved, _ := state.Load(statePath)
	for _, unit := range saved.Sections[section.Slug].Units {
		if unit.Result == "stopped" && unit.UpdatedAt.After(started) {
			return unit.Reason
		}
	}
	return ""
}

// appendEvent adds event to the run file at path; a failed write only prints to stderr, so the log never stops a run.
func appendEvent(path string, event events.Event) {
	if err := events.Append(path, event); err != nil {
		fmt.Fprintln(os.Stderr, "events:", err)
	}
}
