package pr

import (
	"slices"
	"testing"
)

func TestAutoMergeArgsUseRebaseForRebase(t *testing.T) {
	want := []string{"pr", "merge", "--auto", "--rebase", "feat/feed"}
	if args := autoMergeArgs("feat/feed", "rebase"); !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}

func TestAutoMergeArgsUseMergeOtherwise(t *testing.T) {
	want := []string{"pr", "merge", "--auto", "--merge", "feat/feed"}
	if args := autoMergeArgs("feat/feed", "squash"); !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}
