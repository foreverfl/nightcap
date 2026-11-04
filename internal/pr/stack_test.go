package pr

import (
	"slices"
	"testing"
)

func TestStackArgsLinkThePreviousPRUnderTheNewOne(t *testing.T) {
	args := stackArgs("https://github.com/foreverfl/feed/pull/1", "https://github.com/foreverfl/feed/pull/2")
	want := []string{"stack", "link", "https://github.com/foreverfl/feed/pull/1", "https://github.com/foreverfl/feed/pull/2"}
	if !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}
