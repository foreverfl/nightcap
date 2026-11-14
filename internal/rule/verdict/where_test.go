package verdict

import (
	"slices"
	"strings"
	"testing"
)

func TestWherePathsDropsNotesAndBackticks(t *testing.T) {
	body := strings.Join([]string{
		"- [ ] add the feed",
		"  - WHAT  list it",
		"  - WHERE internal/a.go, internal/b.go(설명) · `c.go`",
	}, "\n")
	want := []string{"internal/a.go", "internal/b.go", "c.go"}
	if paths := WherePaths(body); !slices.Equal(paths, want) {
		t.Fatalf("want %q, got %q", want, paths)
	}
}

func TestWherePathsWithoutWhereLineIsNil(t *testing.T) {
	body := strings.Join([]string{
		"- [ ] add the feed",
		"  - WHAT  list it",
	}, "\n")
	if paths := WherePaths(body); paths != nil {
		t.Fatalf("a body with no WHERE line should give nil, got %q", paths)
	}
}

func TestWherePathsExpandsBraces(t *testing.T) {
	body := strings.Join([]string{
		"- [ ] add the feed",
		"  - WHERE internal/check/{run.go, classify.go}, main.go",
	}, "\n")
	want := []string{"internal/check/run.go", "internal/check/classify.go", "main.go"}
	if paths := WherePaths(body); !slices.Equal(paths, want) {
		t.Fatalf("want %q, got %q", want, paths)
	}
}
