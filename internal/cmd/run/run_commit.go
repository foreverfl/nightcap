package run

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"nightcap/internal/client/claude"
	"nightcap/internal/client/git"
	"nightcap/internal/client/test"
	"nightcap/internal/rule/usage"
	"nightcap/internal/rule/verdict"
	"nightcap/internal/store/events"
	"nightcap/internal/store/plan"
	"nightcap/internal/store/state"
	"nightcap/internal/ui/indent"
)

// runCommit runs claude and the test on one commit until the test is green or MaxRounds is spent.
// It prints "blocked" when the rounds run out, the same failure comes back twice, or claude stops before Done.
func runCommit(ctx context.Context, stdout io.Writer, nightRun nightRunSettings, worktreeDir, base, slug string, number int, commit plan.Commit) error {
	failureTail := ""
	sessionID := ""
	record := state.Commit{}
	appendEvent(nightRun.runPath, events.Event{Kind: "commit_start", At: time.Now(), PR: slug, Commit: commit.Title})
	for round := 1; round <= nightRun.loaded.MaxRounds; round++ {
		record.Attempts = round
		logPath := filepath.Join(filepath.Dir(nightRun.runPath), slug, fmt.Sprintf("commit%d-round%d.jsonl", number, round))
		roundStarted := time.Now()
		result, err := claude.RunRound(ctx, claude.Round{
			WorktreeDir:  worktreeDir,
			Prompt:       BuildPrompt(commit.Body, failureTail, sessionID != ""),
			SessionID:    sessionID,
			MaxTurns:     nightRun.loaded.MaxTurns,
			AllowedTools: nightRun.loaded.AllowedTools,
			AuthorName:   nightRun.loaded.AuthorName,
			AuthorEmail:  nightRun.loaded.AuthorEmail,
			LogPath:      logPath,
		}, stdout)
		if err != nil {
			return err
		}
		if result.SessionID != "" {
			sessionID = result.SessionID
		}
		appendEvent(nightRun.runPath, events.Event{Kind: "round", At: time.Now(), PR: slug, Commit: commit.Title,
			Round: round, Seconds: int(time.Since(roundStarted).Seconds()), SessionID: result.SessionID})
		indent.Fprintf(stdout, 1, "ran  %s: %s (log: %s)\n", commit.Title, result.Outcome, logPath)
		if result.Outcome == claude.RateLimited {
			usageJSON, _ := claude.Usage(ctx)
			current, _ := usage.Read(usageJSON)
			if current.Known {
				record.UsageAtLimit = &state.UsageAtLimit{Session: current.Session.Percent, WeeklyAll: current.WeeklyAll.Percent}
			}
			reset, afterWork := usage.WindowReset(time.Now(), current, nightRun.loaded.WorkStart)
			// the reset comes after work_start, so sleeping would carry the run into the workday
			if afterWork != "" {
				indent.Fprintf(stdout, 1, "stop: rate limited, %s\n", afterWork)
				record.Result, record.Reason = "stopped", "rate limited, "+afterWork
				return state.SaveCommit(nightRun.statePath, slug, commit.Title, record)
			}
			indent.Fprintf(stdout, 1, "sleep until %s\n", reset.Local().Format("15:04"))
			sleepUntil(ctx, reset)
			continue
		}
		if result.Outcome != claude.Done {
			break
		}
		// 1. Test: run the test command from the plan's frontmatter (pass or not)
		tested, err := test.Run(ctx, worktreeDir, nightRun.frontmatter.Test)
		if err != nil {
			return err
		}
		indent.Fprintf(stdout, 1, "test: %s\n", color(tested.Green))
		appendEvent(nightRun.runPath, events.Event{Kind: "test", At: time.Now(), PR: slug, Commit: commit.Title, Result: color(tested.Green)})
		failure := tested.Tail
		record.Result, record.Reason, record.TestTail = "red", "", tested.Tail
		// 2. Verdict: sort what the branch changed by the rules in verdict/classify.go
		if tested.Green {
			verdict, err := classifyBranch(stdout, worktreeDir, base, nightRun.frontmatter, commit)
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
			appendEvent(nightRun.runPath, events.Event{Kind: "verdict", At: time.Now(), PR: slug, Commit: commit.Title, Result: record.Result, Reason: verdict.Reason})
			if !verdict.Red {
				return state.SaveCommit(nightRun.statePath, slug, commit.Title, record)
			}
			failure = verdict.Reason
		}
		if err := state.SaveCommit(nightRun.statePath, slug, commit.Title, record); err != nil {
			return err
		}
		sameFailureAsLastRound := round > 1 && failure == failureTail
		if sameFailureAsLastRound {
			break
		}
		failureTail = failure
	}
	indent.Fprintf(stdout, 1, "blocked\n")
	appendEvent(nightRun.runPath, events.Event{Kind: "blocked", At: time.Now(), PR: slug, Commit: commit.Title})
	record.Result = "blocked"
	return state.SaveCommit(nightRun.statePath, slug, commit.Title, record)
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
