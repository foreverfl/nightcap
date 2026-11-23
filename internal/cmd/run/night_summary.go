package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nightcap/internal/client/claude"
	"nightcap/internal/client/notify"
	"nightcap/internal/rule/usage"
	"nightcap/internal/store/config"
	"nightcap/internal/store/plan"
	"nightcap/internal/store/state"
)

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
