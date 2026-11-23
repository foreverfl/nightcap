package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"nightcap/internal/client/claude"
	"nightcap/internal/client/ghapp"
	"nightcap/internal/client/git"
	"nightcap/internal/client/pr"
	"nightcap/internal/client/release"
	"nightcap/internal/rule/usage"
	"nightcap/internal/store/config"
	"nightcap/internal/store/events"
	"nightcap/internal/store/path"
	"nightcap/internal/store/plan"
	"nightcap/internal/store/state"
	"nightcap/internal/ui/indent"
)

// nightRunSettings holds the values one `nightcap run` reads once and keeps until it ends.
type nightRunSettings struct {
	loaded      config.Config // nightcap.yml
	frontmatter plan.Meta     // the plan file's repo, test, protected and dependencies
	statePath   string        // state.json, where each commit's result is kept across runs
	runPath     string        // this run's events file
}

// Run goes down the plan file args[0] PR by PR, running every open commit in a worktree of the frontmatter's repo, until nothing is left or a commit is not accepted.
// version is this binary's release tag; a newer one on GitHub gets one line before the work starts.
// Cancelling ctx stops claude and the test where they are; the run then ends as it does on Ctrl+C.
// It fails when no plan file is named or the frontmatter leaves a key out, and when the config, git, claude or sh fails.
func Run(ctx context.Context, args []string, version string, stdout io.Writer) (runErr error) {
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
	// one folder per run, named by its start: ~/.local/share/nightcap/blog-v2/runs/261108/225500/events.jsonl
	runPath := filepath.Join(dataDir, "runs", started.Format("060102"), started.Format("150405"), "events.jsonl")
	if err := os.MkdirAll(filepath.Dir(runPath), 0o755); err != nil {
		return err
	}
	runLogFile, err := os.OpenFile(filepath.Join(filepath.Dir(runPath), "run.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer runLogFile.Close()
	stdout = io.MultiWriter(stdout, runLogFile)
	for _, noticeLine := range release.NewVersionNotice(ctx, version) {
		fmt.Fprintln(stdout, noticeLine)
	}
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

	nightRun := nightRunSettings{loaded: loaded, frontmatter: meta, statePath: statePath, runPath: runPath}
	// a PR whose commits all ended accepted lets the next pass take the PR below it
	for {
		picked, finished, err := runNextPR(ctx, stdout, nightRun, planPath, string(planText), worktreesDir, &startUsage)
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
func runNextPR(ctx context.Context, stdout io.Writer, nightRun nightRunSettings, planPath, planText, worktreesDir string, startUsage *usage.Usage) (picked *plan.PR, finished bool, err error) {
	saved, err := state.Load(nightRun.statePath)
	if err != nil {
		return picked, false, err
	}
	var prsToDo []plan.PR
	hasPR := map[string]bool{}
	baseFor := map[string]string{}
	previousOpenBranch := ""
	stackBelow := map[string][]string{} // the open PRs a new PR stacks onto, bottom first
	var openStack []string              // the open PRs right above in the plan, unbroken, bottom first
	for _, planPR := range plan.Parse(planText) {
		baseFor[planPR.Slug] = "main"
		if previousOpenBranch != "" {
			baseFor[planPR.Slug] = previousOpenBranch
		}
		stackBelow[planPR.Slug] = openStack
		previousOpenBranch = ""
		openStack = nil
		if len(planPR.Commits) > 0 {
			view, err := pr.ViewBranch(nightRun.frontmatter.Repo, planPR.Type+"/"+planPR.Slug)
			if err != nil {
				return picked, false, err
			}
			if err := state.SavePRStatus(nightRun.statePath, planPR.Slug, view.URL, view.State); err != nil {
				return picked, false, err
			}
			if view.State == "MERGED" || view.State == "CLOSED" {
				if err := removeWorktreeIfAny(stdout, nightRun.frontmatter.Repo, filepath.Join(worktreesDir, planPR.Slug), planPR.Slug, view.State); err != nil {
					return picked, false, err
				}
			}
			if view.State == "MERGED" {
				fmt.Fprintf(stdout, "skip %s: PR merged\n", planPR.Title)
				continue
			}
			if view.State == "CLOSED" {
				if err := git.DeleteBranch(nightRun.frontmatter.Repo, planPR.Type+"/"+planPR.Slug); err != nil {
					return picked, false, err
				}
				if _, err := git.DeleteRemoteBranch(nightRun.frontmatter.Repo, planPR.Type+"/"+planPR.Slug); err != nil {
					return picked, false, err
				}
				if err := state.ForgetPR(nightRun.statePath, planPR.Slug); err != nil {
					return picked, false, err
				}
				fmt.Fprintf(stdout, "redo %s: PR closed\n", planPR.Title)
				prsToDo = append(prsToDo, planPR)
				continue
			}
			hasPR[planPR.Slug] = view.State == "OPEN"
			if hasPR[planPR.Slug] {
				previousOpenBranch = planPR.Type + "/" + planPR.Slug
				openStack = append(slices.Clone(stackBelow[planPR.Slug]), view.URL)
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
	worktreeDir, err := git.AddWorktree(nightRun.frontmatter.Repo, base, branch, filepath.Join(worktreesDir, planPR.Slug))
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
		if done, _ := isAccepted(nightRun.statePath, planPR.Slug, commit.Title); done {
			continue // done in an earlier run; the plan still shows it as [ ]
		}
		// a failed /usage reads as unknown usage; the percent limits are skipped and the rate-limit error is trusted instead
		usageJSON, _ := claude.Usage(ctx)
		current, _ := usage.Read(usageJSON)
		if !startUsage.Known {
			*startUsage = current
		}
		if reason := usageStopReason(time.Now(), current, nightRun.loaded); reason != "" {
			indent.Fprintf(stdout, 1, "stop: %s\n", reason)
			return picked, false, state.SaveCommit(nightRun.statePath, planPR.Slug, commit.Title, state.Commit{Result: "stopped", Reason: reason})
		}
		if err := runCommit(ctx, stdout, nightRun, worktreeDir, base, planPR.Slug, index+1, commit); err != nil {
			return picked, false, err
		}
		accepted, err := isAccepted(nightRun.statePath, planPR.Slug, commit.Title)
		if err != nil {
			return picked, false, err
		}
		if !accepted {
			return picked, false, nil // red, blocked or stopped: the commits below stand on this one
		}
		if err := openPR(ctx, stdout, nightRun, worktreeDir, base, stackBelow[planPR.Slug], planPR, commit.Title, hasPR[planPR.Slug]); err != nil {
			return picked, false, err
		}
		hasPR[planPR.Slug] = true
		if err := labelWhenFinished(ctx, stdout, nightRun.statePath, worktreeDir, planPR, nightRun.loaded); err != nil {
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
