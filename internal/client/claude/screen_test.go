package claude

import (
	"slices"
	"strings"
	"testing"
)

const worktreeDir = "/home/me/.cache/nightcap/blog-v2/worktrees/feed"

func TestScreenLinesShowsAnEditWithTheFileUnderTheWorktree(t *testing.T) {
	raw := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Edit","input":{"file_path":"` + worktreeDir + `/internal/feed/feed.go"}}]}}`

	got := ScreenLines([]byte(raw), worktreeDir)

	if !slices.Equal(got, []string{"Edit internal/feed/feed.go"}) {
		t.Fatalf("got %q", got)
	}
}

func TestScreenLinesShowsABashCommand(t *testing.T) {
	raw := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"go test ./internal/feed"}}]}}`

	got := ScreenLines([]byte(raw), worktreeDir)

	if !slices.Equal(got, []string{"Bash go test ./internal/feed"}) {
		t.Fatalf("got %q", got)
	}
}

func TestScreenLinesCutsTextToItsFirst80Characters(t *testing.T) {
	longFirstLine := strings.Repeat("가", 100)
	raw := `{"type":"assistant","message":{"content":[{"type":"text","text":"` + longFirstLine + `\nsecond line"}]}}`

	got := ScreenLines([]byte(raw), worktreeDir)

	if !slices.Equal(got, []string{`"` + strings.Repeat("가", 80) + `"`}) {
		t.Fatalf("got %q", got)
	}
}

func TestScreenLinesShowsNothingForAResultLine(t *testing.T) {
	raw := `{"type":"result","result":"ok"}`

	got := ScreenLines([]byte(raw), worktreeDir)

	if len(got) != 0 {
		t.Fatalf("got %q", got)
	}
}
