package usage

import (
	"testing"
	"time"
)

func TestNextWorkStart(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	night := time.Date(2026, 11, 5, 23, 10, 0, 0, tokyo)
	morning := time.Date(2026, 11, 5, 8, 0, 0, 0, tokyo)

	if got, ok := NextWorkStart(night, "09:00"); !ok || !got.Equal(time.Date(2026, 11, 6, 9, 0, 0, 0, tokyo)) {
		t.Fatalf("23:10 → %v, %v; want 11/6 09:00", got, ok)
	}
	if got, ok := NextWorkStart(morning, "09:00"); !ok || !got.Equal(time.Date(2026, 11, 5, 9, 0, 0, 0, tokyo)) {
		t.Fatalf("08:00 → %v, %v; want 11/5 09:00", got, ok)
	}
	if _, ok := NextWorkStart(night, ""); ok {
		t.Fatal("empty work_start → want false")
	}
}
