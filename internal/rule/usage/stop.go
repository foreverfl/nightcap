package usage

import (
	"fmt"
	"time"
)

// StopReason says why no commit should start now under limitSession and limitWeeklyAll (0 = no limit), or "" when one may.
func StopReason(now time.Time, current Usage, limitSession, limitWeeklyAll int, workStartHHMM string) string {
	// stop 1: the 5h window is already at limit_session, so this commit could hit the limit halfway
	if current.Known && limitSession > 0 && current.Session.Percent >= limitSession {
		return fmt.Sprintf("session %d%% ≥ %d", current.Session.Percent, limitSession)
	}
	// stop 2: the week is at limit_weekly_all and its reset is over 24h away, so tonight would eat the weekdays' share
	weeklyResetsSoon := current.WeeklyAll.ResetsAt.Sub(now) < 24*time.Hour
	if current.Known && limitWeeklyAll > 0 && !weeklyResetsSoon && current.WeeklyAll.Percent >= limitWeeklyAll {
		return fmt.Sprintf("weekly_all %d%% ≥ %d", current.WeeklyAll.Percent, limitWeeklyAll)
	}
	// stop 3: the window would still be running at work_start, so the morning would start rate-limited
	_, afterWork := WindowReset(now, current, workStartHHMM)
	return afterWork
}

// WindowReset returns when the 5h window resets, and afterWork such as "reset 10:00 > work_start 09:00" when that is after work_start.
func WindowReset(now time.Time, current Usage, workStartHHMM string) (reset time.Time, afterWork string) {
	reset = now.Add(5 * time.Hour)
	if current.Session.ResetsAt.After(now) {
		reset = current.Session.ResetsAt
	}
	workStart, _ := NextWorkStart(now, workStartHHMM)
	if EndsBeforeWork(reset, workStart) {
		return reset, ""
	}
	return reset, fmt.Sprintf("reset %s > work_start %s", reset.In(now.Location()).Format("15:04"), workStart.Format("15:04"))
}
