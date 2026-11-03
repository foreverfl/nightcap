package pr

import (
	"slices"
	"testing"
)

func TestDraftArgsOpenADraftWithTitleAndBodyFile(t *testing.T) {
	args := draftArgs("main", "feat: feed", "/tmp/body.md")
	want := []string{"pr", "create", "--draft", "--base", "main", "--title", "feat: feed", "--body-file", "/tmp/body.md"}
	if !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}
