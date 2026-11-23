package usage

import (
	"testing"
	"time"
)

func TestStopReason(t *testing.T) {
	now := time.Date(2026, 11, 5, 23, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		current Usage
		want    string
	}{
		// stops at stop 1: session 85% is at or above limit_session 80
		{"session over", Usage{Known: true,
			Session:   Limit{Percent: 85},
			WeeklyAll: Limit{Percent: 30, ResetsAt: now.Add(72 * time.Hour)}}, "session 85% ≥ 80"},
		// stops at stop 2: weekly 62% is at or above limit_weekly_all 60 and its reset is 3 days away
		{"weekly over, reset in 3 days", Usage{Known: true,
			Session:   Limit{Percent: 30},
			WeeklyAll: Limit{Percent: 62, ResetsAt: now.Add(72 * time.Hour)}}, "weekly_all 62% ≥ 60"},
		// opens: weekly 62% is over 60, but the reset is 9h away, so the unused share would vanish anyway
		{"weekly over, reset in 9h", Usage{Known: true,
			Session:   Limit{Percent: 30},
			WeeklyAll: Limit{Percent: 62, ResetsAt: now.Add(9 * time.Hour)}}, ""},
		// opens: both limits are under their caps, and work_start is empty so stop 3 never applies
		{"both at 30%", Usage{Known: true,
			Session:   Limit{Percent: 30},
			WeeklyAll: Limit{Percent: 30, ResetsAt: now.Add(72 * time.Hour)}}, ""},
	}
	for _, c := range cases {
		if got := StopReason(now, c.current, 80, 60, ""); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
