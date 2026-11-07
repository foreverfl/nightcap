package run

import (
	"context"
	"flag"
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
	"nightcap/internal/git"
	"nightcap/internal/notify"
	"nightcap/internal/plan"
	"nightcap/internal/pr"
	"nightcap/internal/state"
	"nightcap/internal/usage"
)

// Run is the "run" subcommand; args is everything after "run", where --section names one heading.
// For now it picks tonight's heading per target, checks its branch out in a worktree and runs
// the first unit that has a check through runUnit; it returns an error when the config,
// a plan file, git, claude or sh fails.
func Run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	sectionName := flags.String("section", "", "heading title or slug to take")
	if err := flags.Parse(args); err != nil {
		return err
	}
	loaded, err := config.Load(".")
	if err != nil {
		return err
	}
	defer sendNightSummary(loaded, time.Now())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for _, target := range loaded.Targets {
		planText, err := os.ReadFile(target.Plan)
		if err != nil {
			return err
		}
		statePath := filepath.Join(target.Repo, ".nightcap", "state.json")
		saved, err := state.Load(statePath)
		if err != nil {
			return err
		}
		var notBlocked []plan.Section
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
				fmt.Fprintf(stdout, "skip %s: blocked\n", section.Title)
				continue
			}
			if len(section.Units) > 0 {
				view, err := pr.ViewBranch(target.Repo, section.Type+"/"+section.Slug)
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
		section, ok := plan.Pick(notBlocked, *sectionName)
		if !ok {
			fmt.Fprintf(stdout, "%s: nothing to take\n", target.Plan)
			continue
		}
		base := baseFor[section.Slug]
		branch := section.Type + "/" + section.Slug
		worktreeDir, err := git.AddWorktree(target.Repo, base, branch, section.Slug)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "%s\n  branch: %s\n  worktree: %s\n", target.Plan, branch, worktreeDir)
		for _, unit := range section.Units {
			mark := "skip"
			if unit.HasCheck {
				mark = "take"
			}
			fmt.Fprintf(stdout, "  %s  %s\n", mark, unit.Title)
		}

		for index, unit := range section.Units {
			if !unit.HasCheck {
				continue
			}
			// a failed /usage reads as unknown usage; the percent limits are skipped and the rate-limit error is trusted instead
			usageJSON, _ := claude.Usage(ctx)
			current, _ := usage.Read(usageJSON)
			if reason := usageStopReason(time.Now(), current, loaded); reason != "" {
				fmt.Fprintf(stdout, "  stop: %s\n", reason)
				return state.SaveUnit(statePath, section.Slug, unit.Title, state.Unit{Result: "stopped", Reason: reason})
			}
			if err := runUnit(ctx, stdout, loaded, target, worktreeDir, base, section.Slug, index+1, unit, statePath); err != nil {
				return err
			}
			if err := openDraftPR(ctx, stdout, statePath, worktreeDir, base, stackOnto[section.Slug], section, unit.Title, hasPR[section.Slug], slices.Contains(target.Lang, "en")); err != nil {
				return err
			}
			if err := markReadyWhenFinished(stdout, statePath, worktreeDir, target.MergeMethod, section); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

// runUnit runs claude then the check on one unit up to MaxRounds times, feeding a red check's tail or a red verdict's
// reason into the next prompt. It prints "blocked" when the rounds run out, the same failure comes back twice in a row,
// or claude stops before Done. Every round's outcome goes into the state file at statePath.
// It returns an error only when claude, sh, git or the state file fails.
func runUnit(ctx context.Context, stdout io.Writer, loaded config.Config, target config.Target, worktreeDir, base, slug string, number int, unit plan.Unit, statePath string) error {
	failureTail := ""
	sessionID := ""
	record := state.Unit{}
	for round := 1; round <= loaded.MaxRounds; round++ {
		record.Attempts = round
		logPath := filepath.Join(target.Repo, ".nightcap", "logs", slug, fmt.Sprintf("unit%d-round%d.log", number, round))
		result, err := claude.RunRound(ctx, worktreeDir, BuildPrompt(unit.Body, failureTail, sessionID != ""), sessionID, loaded.MaxTurns, loaded.AllowedTools, logPath)
		if err != nil {
			return err
		}
		if result.SessionID != "" {
			sessionID = result.SessionID
		}
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
		checked, err := check.Run(ctx, worktreeDir, target.Check)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "  check: %s\n", color(checked.Green))
		failure := checked.Tail
		record.Result, record.Reason, record.CheckTail = "red", "", checked.Tail
		// 2. Classify: sort what the branch changed by the rules in check/classify.go
		if checked.Green {
			verdict, err := classifyBranch(stdout, worktreeDir, base, target, unit)
			if err != nil {
				return err
			}
			record.Reason = verdict.Reason
			if !verdict.Red {
				record.Result = "green"
				if verdict.ReadCode {
					record.Result = "read-code"
				}
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
func classifyBranch(stdout io.Writer, worktreeDir, base string, target config.Target, unit plan.Unit) (check.Verdict, error) {
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
	verdict := check.Classify(changes, addedLines, check.WherePaths(unit.Body), target.ReadCode, target.Protected, target.Dependencies)
	if verdict.Red {
		fmt.Fprintf(stdout, "  verdict: red (%s)\n", verdict.Reason)
	} else if verdict.ReadCode {
		fmt.Fprintf(stdout, "  verdict: green, read the code (%s)\n", verdict.Reason)
	} else {
		fmt.Fprintln(stdout, "  verdict: green")
	}
	return verdict, nil
}

// openDraftPR pushes section's branch from worktreeDir once unitTitle has ended green or read-code in the state file at statePath,
// then opens a draft PR, or with hasPR replaces its description, listing every finished unit (English above Korean when
// isEnglishWanted). A new PR on top of the open PR at previousURL joins its stack. It returns an error when git, gh or the state file fails.
func openDraftPR(ctx context.Context, stdout io.Writer, statePath, worktreeDir, base, previousURL string, section plan.Section, unitTitle string, hasPR, isEnglishWanted bool) error {
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
	prURL, err := pr.CreateDraft(worktreeDir, base, section.Type+": "+section.Title, body)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "  draft PR: %s\n", prURL)
	if previousURL != "" {
		if err := pr.LinkStack(worktreeDir, previousURL, prURL); err != nil {
			return err
		}
	}
	return state.SavePRStatus(statePath, section.Slug, prURL, "OPEN")
}

// bodyInput collects the PR description's values: section's summary, every unit of section that ended green or read-code
// in sectionState, in plan order, and the last check tail of lastUnitTitle.
func bodyInput(section plan.Section, sectionState state.Section, lastUnitTitle string) pr.BodyInput {
	input := pr.BodyInput{Summary: section.Summary, CheckTail: sectionState.Units[lastUnitTitle].CheckTail}
	for _, unit := range section.Units {
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

// markReadyWhenFinished marks the PR ready when every unit is done, adds the label the heading's mark names, and turns on auto-merge for 🤖.
func markReadyWhenFinished(stdout io.Writer, statePath, worktreeDir, mergeMethod string, section plan.Section) error {
	saved, err := state.Load(statePath)
	if err != nil {
		return err
	}
	for _, unit := range section.Units {
		result := saved.Sections[section.Slug].Units[unit.Title].Result
		if result != "green" && result != "read-code" {
			return nil
		}
	}
	branch := section.Type + "/" + section.Slug
	if err := pr.MarkReady(worktreeDir, branch); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "  ready: %s\n", branch)
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

// sendNightSummary sends one line with the target repos and the start and end times; a failed send only prints to stderr.
func sendNightSummary(loaded config.Config, started time.Time) {
	var repos []string
	for _, target := range loaded.Targets {
		repos = append(repos, filepath.Base(target.Repo))
	}
	text := fmt.Sprintf("nightcap %s · %s → %s", strings.Join(repos, ", "), started.Format("15:04"), time.Now().Format("15:04"))
	if err := notify.Send(loaded.Webhook, text); err != nil {
		fmt.Fprintln(os.Stderr, "notify:", err)
	}
}
