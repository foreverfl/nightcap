package pr

import (
	"slices"
	"testing"
)

func TestCreateArgsOpenAPRWithTitleAndBodyFile(t *testing.T) {
	args := createArgs("main", "feat: feed", "/tmp/body.md")
	want := []string{"pr", "create", "--base", "main", "--title", "feat: feed", "--body-file", "/tmp/body.md"}
	if !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}
