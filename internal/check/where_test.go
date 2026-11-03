package check

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
