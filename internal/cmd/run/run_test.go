package run

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nightcap/internal/config"
	"nightcap/internal/notify"
	"nightcap/internal/plan"
	"nightcap/internal/state"
	"nightcap/internal/usage"
)

func TestUsageStopReason(t *testing.T) {
	now := time.Date(2026, 11, 5, 23, 0, 0, 0, time.UTC)
	limits := config.Config{LimitSession: 80, LimitWeeklyAll: 60}
	cases := []struct {
		name    string
		current usage.Usage
		want    string
	}{
		// stops at stop 1: session 85% is at or above limit_session 80
		{"session over", usage.Usage{Known: true,
			Session:   usage.Limit{Percent: 85},
			WeeklyAll: usage.Limit{Percent: 30, ResetsAt: now.Add(72 * time.Hour)}}, "session 85% ≥ 80"},
		// stops at stop 2: weekly 62% is at or above limit_weekly_all 60 and its reset is 3 days away
		{"weekly over, reset in 3 days", usage.Usage{Known: true,
			Session:   usage.Limit{Percent: 30},
			WeeklyAll: usage.Limit{Percent: 62, ResetsAt: now.Add(72 * time.Hour)}}, "weekly_all 62% ≥ 60"},
		// opens: weekly 62% is over 60, but the reset is 9h away, so the unused share would vanish anyway
		{"weekly over, reset in 9h", usage.Usage{Known: true,
			Session:   usage.Limit{Percent: 30},
			WeeklyAll: usage.Limit{Percent: 62, ResetsAt: now.Add(9 * time.Hour)}}, ""},
		// opens: both limits are under their caps, and work_start is empty so stop 3 never applies
		{"both at 30%", usage.Usage{Known: true,
			Session:   usage.Limit{Percent: 30},
			WeeklyAll: usage.Limit{Percent: 30, ResetsAt: now.Add(72 * time.Hour)}}, ""},
	}
	for _, c := range cases {
		if got := usageStopReason(now, c.current, limits); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestRunCommitSleepsUntilTheResetAfterARateLimit(t *testing.T) {
	repo := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main", repo},
		{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	// the fake answers /usage with session 100% resetting in 2099 and weekly_all 62%, hits the limit on its first round and finishes on the second
	rounds := filepath.Join(t.TempDir(), "rounds")
	fakeClaude := filepath.Join(t.TempDir(), "claude")
	script := strings.Join([]string{
		"#!/bin/sh",
		`if [ "$2" = "/usage" ]; then`,
		`  echo '{"usage_report":{"rate_limits":{"limits":[{"kind":"session","percent":100,"resets_at":"2099-01-01T04:30:00+09:00"},{"kind":"weekly_all","percent":62}]}}}'`,
		"  exit 0",
		"fi",
		"echo x >> " + rounds,
		`if [ "$(wc -l < ` + rounds + `)" -eq 1 ]; then`,
		`  echo '{"type":"result","result":"limit","api_error_status":429}'`,
		"else",
		`  echo '{"type":"result","result":"ok"}'`,
		"fi",
	}, "\n")
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)
	var sleptUntil time.Time
	realSleep := sleepUntil
	t.Cleanup(func() { sleepUntil = realSleep })
	sleepUntil = func(_ context.Context, wakeAt time.Time) { sleptUntil = wakeAt }

	statePath := filepath.Join(repo, ".nightcap", "state.json")
	loaded := config.Config{MaxRounds: 3}
	meta := plan.Meta{Repo: repo, Test: "true"}
	commit := plan.Commit{Title: "add it"}
	if err := runCommit(context.Background(), io.Discard, loaded, meta, repo, "main", "feed", 1, commit, statePath, filepath.Join(repo, ".nightcap", "runs", "run.jsonl")); err != nil {
		t.Fatal(err)
	}
	saved, _ := state.Load(statePath)
	got := saved.PRs["feed"].Commits["add it"]
	reset := time.Date(2099, 1, 1, 4, 30, 0, 0, time.FixedZone("JST", 9*60*60))
	if got.Attempts != 2 || got.Result != "green" || !sleptUntil.Equal(reset) {
		t.Fatalf("got %+v, slept until %v; want 2 rounds, green, slept until %v", got, sleptUntil, reset)
	}
	if got.UsageAtLimit == nil || *got.UsageAtLimit != (state.UsageAtLimit{Session: 100, WeeklyAll: 62}) {
		t.Fatalf("usage_at_limit = %+v, want session 100 and weekly_all 62", got.UsageAtLimit)
	}
}

func TestBuildPromptLeavesTheRulesOutOfAResumedSession(t *testing.T) {
	if !strings.Contains(BuildPrompt("add it", "", false), commitRules) {
		t.Fatal("a new session should get the commit rules")
	}
	if strings.Contains(BuildPrompt("add it", "FAIL", true), commitRules) {
		t.Fatal("a resumed session already has the commit rules")
	}
}

func TestPRSummaryCountsGreenCommitsAndTakesThePRLink(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	if err := state.SaveCommit(statePath, "feed-hurl", "list feeds", state.Commit{Result: "green"}); err != nil {
		t.Fatal(err)
	}
	if err := state.SavePRStatus(statePath, "feed-hurl", "https://github.com/foreverfl/blog-v2/pull/41", "OPEN"); err != nil {
		t.Fatal(err)
	}
	planPR := plan.PR{Mark: plan.MarkAuto, Type: "test", Title: "피드 hurl 파일", Slug: "feed-hurl",
		Commits: []plan.Commit{{Title: "list feeds"}, {Title: "add a feed"}, {Title: "delete a feed"}}}
	got, err := prSummary(statePath, planPR)
	if err != nil {
		t.Fatal(err)
	}
	want := notify.PR{Mark: plan.MarkAuto, Title: "test: 피드 hurl 파일 (feed-hurl)", Done: 1, Total: 3, URL: "https://github.com/foreverfl/blog-v2/pull/41"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestStopReasonReportsOnlyTonightsStop(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	started := time.Date(2026, 11, 8, 23, 0, 0, 0, time.UTC)
	planPR := plan.PR{Slug: "feed-hurl", Commits: []plan.Commit{{Title: "list feeds"}}}
	for _, tc := range []struct {
		updatedAt time.Time
		want      string
	}{
		{started.Add(time.Hour), "session 85% ≥ 80"},
		{started.Add(-24 * time.Hour), ""},
	} {
		saved := state.State{PRs: map[string]state.PR{"feed-hurl": {Commits: map[string]state.Commit{
			"list feeds": {Result: "stopped", Reason: "session 85% ≥ 80", UpdatedAt: tc.updatedAt},
		}}}}
		if err := state.Save(statePath, saved); err != nil {
			t.Fatal(err)
		}
		if got := stopReason(statePath, planPR, started); got != tc.want {
			t.Fatalf("stopped at %v: got %q, want %q", tc.updatedAt, got, tc.want)
		}
	}
}
