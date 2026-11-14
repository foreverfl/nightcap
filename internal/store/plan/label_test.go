package plan

import (
	"strings"
	"testing"
)

func TestLabelValueReadsWhatAndDone(t *testing.T) {
	body := strings.Join([]string{
		"- [ ] add the feed",
		"  - WHAT  list the feed",
		"  - DONE  go test ./internal/feed passes",
	}, "\n")
	if what := LabelValue(body, "WHAT"); what != "list the feed" {
		t.Fatalf("WHAT should be %q, got %q", "list the feed", what)
	}
	if done := LabelValue(body, "DONE"); done != "go test ./internal/feed passes" {
		t.Fatalf("DONE should be %q, got %q", "go test ./internal/feed passes", done)
	}
}

func TestLabelValueWithoutDoneIsEmpty(t *testing.T) {
	body := strings.Join([]string{
		"- [ ] add the feed",
		"  - WHAT  list the feed",
	}, "\n")
	if done := LabelValue(body, "DONE"); done != "" {
		t.Fatalf("a body with no DONE line should give an empty string, got %q", done)
	}
}
