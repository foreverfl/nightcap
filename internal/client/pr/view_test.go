package pr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// useFakeGh points NIGHTCAP_GH at a shell script made of lines.
func useFakeGh(t *testing.T, lines ...string) {
	t.Helper()
	fakeGh := filepath.Join(t.TempDir(), "gh")
	script := strings.Join(append([]string{"#!/bin/sh"}, lines...), "\n")
	if err := os.WriteFile(fakeGh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_GH", fakeGh)
}

func TestViewBranchReadsAMergedPR(t *testing.T) {
	useFakeGh(t, `echo '{"state":"MERGED","url":"https://github.com/foreverfl/feed/pull/1"}'`)
	view, err := ViewBranch(t.TempDir(), "feat/feed")
	if err != nil {
		t.Fatal(err)
	}
	if view != (View{State: "MERGED", URL: "https://github.com/foreverfl/feed/pull/1"}) {
		t.Fatalf("got %+v", view)
	}
}

func TestViewBranchWithoutPRIsEmpty(t *testing.T) {
	useFakeGh(t, `echo 'no pull requests found for branch "feat/feed"' >&2`, "exit 1")
	view, err := ViewBranch(t.TempDir(), "feat/feed")
	if err != nil {
		t.Fatal(err)
	}
	if view != (View{}) {
		t.Fatalf("a branch with no PR should give an empty View, got %+v", view)
	}
}
