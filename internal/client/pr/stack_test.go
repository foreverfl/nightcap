package pr

import (
	"slices"
	"testing"
)

func TestStackArgsPassEveryPRFromTheBottom(t *testing.T) {
	args := stackArgs([]string{
		"https://github.com/foreverfl/feed/pull/1",
		"https://github.com/foreverfl/feed/pull/2",
		"https://github.com/foreverfl/feed/pull/3",
	})
	want := []string{
		"stack", "link",
		"https://github.com/foreverfl/feed/pull/1",
		"https://github.com/foreverfl/feed/pull/2",
		"https://github.com/foreverfl/feed/pull/3",
	}
	if !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}
