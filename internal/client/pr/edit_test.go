package pr

import (
	"slices"
	"testing"
)

func TestEditArgsReplaceTheBodyOfTheBranchPR(t *testing.T) {
	args := editArgs("feat/feed", "/tmp/body.md")
	want := []string{"pr", "edit", "feat/feed", "--body-file", "/tmp/body.md"}
	if !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}
