package usage

import "time"

// NextWorkStart turns work_start "HH:MM" into its next occurrence after now, in now's Location; false when hhmm is empty or not HH:MM.
func NextWorkStart(now time.Time, hhmm string) (time.Time, bool) {
	clock, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, false
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), clock.Hour(), clock.Minute(), 0, 0, now.Location())
	if !start.After(now) {
		start = start.AddDate(0, 0, 1)
	}
	return start, true
}
