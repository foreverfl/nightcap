package usage

import "time"

// EndsBeforeWork reports whether a window resetting at reset is over by workStart; a zero workStart (work_start empty) always fits.
func EndsBeforeWork(reset, workStart time.Time) bool {
	return workStart.IsZero() || !reset.After(workStart)
}
