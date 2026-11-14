package pr

import (
	"slices"
	"testing"

	"nightcap/internal/client/git"
)

func TestRiskLinesNameTheFileThatReadsEnv(t *testing.T) {
	added := []git.AddedLine{{Path: "internal/feed/client.go", Text: `	apiKey := os.Getenv("FEED_API_KEY")`}}
	want := []string{"env: internal/feed/client.go"}
	if risks := RiskLines(nil, added); !slices.Equal(risks, want) {
		t.Fatalf("want %q, got %q", want, risks)
	}
}

func TestRiskLinesAreEmptyWithoutRiskyLines(t *testing.T) {
	added := []git.AddedLine{{Path: "internal/feed/feed.go", Text: "	return items"}}
	if risks := RiskLines(nil, added); len(risks) != 0 {
		t.Fatalf("a diff with nothing to watch should give no risk lines, got %q", risks)
	}
}
