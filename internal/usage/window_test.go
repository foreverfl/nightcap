package usage

import (
	"testing"
	"time"
)

func TestEndsBeforeWork(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	work := time.Date(2026, 11, 6, 9, 0, 0, 0, tokyo)
	early := time.Date(2026, 11, 6, 4, 30, 0, 0, tokyo)
	late := time.Date(2026, 11, 6, 10, 0, 0, 0, tokyo)

	if !EndsBeforeWork(early, work) {
		t.Fatal("reset 04:30, work 09:00 → want true")
	}
	if EndsBeforeWork(late, work) {
		t.Fatal("reset 10:00, work 09:00 → want false")
	}
	if !EndsBeforeWork(late, time.Time{}) {
		t.Fatal("reset 10:00, no work start → want true")
	}
}
