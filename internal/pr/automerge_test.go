package pr

import (
	"slices"
	"testing"
)

func TestAutoMergeArgsNameTheBranch(t *testing.T) {
	want := []string{"pr", "merge", "--auto", "--merge", "feat/feed"}
	if args := autoMergeArgs("feat/feed"); !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}
