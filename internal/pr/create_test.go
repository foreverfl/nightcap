package pr

import (
	"slices"
	"testing"
)

func TestCreateArgsOpenADraftWithTitleAndBodyFile(t *testing.T) {
	args := CreateArgs("feat: feed", "/tmp/body.md")
	want := []string{"pr", "create", "--draft", "--title", "feat: feed", "--body-file", "/tmp/body.md"}
	if !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}
