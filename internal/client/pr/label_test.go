package pr

import (
	"testing"

	"nightcap/internal/store/plan"
)

func TestReadingLabelFollowsTheMark(t *testing.T) {
	want := map[plan.Mark]string{
		plan.MarkAuto:   "nightcap:auto",
		plan.MarkReadPR: "nightcap:read-pr",
		plan.MarkNone:   "nightcap:read-pr",
	}
	for mark, label := range want {
		if got := ReadingLabel(mark); got != label {
			t.Fatalf("%s: got %q, want %q", mark, got, label)
		}
	}
}

func TestLabelColorIsGreenForAutoAndPurpleForReadPR(t *testing.T) {
	want := map[string]string{
		"nightcap:auto":    "0E8A16",
		"nightcap:read-pr": "7057FF",
	}
	for label, color := range want {
		if got := labelColor(label); got != color {
			t.Errorf("%s: got %q, want %q", label, got, color)
		}
	}
}

func TestLabelDescriptionSaysHowEachLabelMerges(t *testing.T) {
	want := map[string]string{
		"nightcap:auto":    "merges by itself once the checks pass",
		"nightcap:read-pr": "read the PR body, then merge",
	}
	for label, description := range want {
		if got := labelDescription(label); got != description {
			t.Errorf("%s: got %q, want %q", label, got, description)
		}
	}
}
