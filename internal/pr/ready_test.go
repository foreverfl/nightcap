package pr

import (
	"slices"
	"testing"
)

func TestReadyArgsNameTheBranch(t *testing.T) {
	want := []string{"pr", "ready", "feat/feed"}
	if args := readyArgs("feat/feed"); !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}
