package pr

import (
	"testing"

	"nightcap/internal/plan"
)

func TestReadingLabelFollowsTheMark(t *testing.T) {
	want := map[plan.Mark]string{
		plan.MarkAuto:     "nightcap:auto",
		plan.MarkReadPR:   "nightcap:read-pr",
		plan.MarkReadCode: "nightcap:read-code",
	}
	for mark, label := range want {
		if got := ReadingLabel(mark); got != label {
			t.Fatalf("%s: got %q, want %q", mark, got, label)
		}
	}
}
