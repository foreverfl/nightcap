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

func TestRunUnitSleepsUntilTheResetAfterARateLimit(t *testing.T) {
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
	target := config.Target{Repo: repo, Check: "true"}
	unit := plan.Unit{Title: "add it"}
	if err := runUnit(context.Background(), io.Discard, loaded, target, repo, "main", "feed", 1, unit, statePath); err != nil {
		t.Fatal(err)
	}
	saved, _ := state.Load(statePath)
	got := saved.Sections["feed"].Units["add it"]
	reset := time.Date(2099, 1, 1, 4, 30, 0, 0, time.FixedZone("JST", 9*60*60))
	if got.Attempts != 2 || got.Result != "green" || !sleptUntil.Equal(reset) {
		t.Fatalf("got %+v, slept until %v; want 2 rounds, green, slept until %v", got, sleptUntil, reset)
	}
	if got.UsageAtLimit == nil || *got.UsageAtLimit != (state.UsageAtLimit{Session: 100, WeeklyAll: 62}) {
		t.Fatalf("usage_at_limit = %+v, want session 100 and weekly_all 62", got.UsageAtLimit)
	}
}

func TestBuildPromptLeavesTheRulesOutOfAResumedSession(t *testing.T) {
	if !strings.Contains(BuildPrompt("add it", "", false), unitRules) {
		t.Fatal("a new session should get the unit rules")
	}
	if strings.Contains(BuildPrompt("add it", "FAIL", true), unitRules) {
		t.Fatal("a resumed session already has the unit rules")
	}
}
