package pr

import "testing"

func TestReadingLabelWithoutReadCodeIsReadPR(t *testing.T) {
	if label := ReadingLabel([]string{"green", "green"}); label != "nightcap:read-pr" {
		t.Fatalf("all green should be nightcap:read-pr, got %q", label)
	}
}

func TestReadingLabelWithOneReadCodeIsReadCode(t *testing.T) {
	if label := ReadingLabel([]string{"green", "read-code"}); label != "nightcap:read-code" {
		t.Fatalf("one read-code should be nightcap:read-code, got %q", label)
	}
}
