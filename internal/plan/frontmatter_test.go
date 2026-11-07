package plan

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// withFrontmatter is a one-heading plan under a frontmatter naming every Meta field.
const withFrontmatter = `---
repo: ~/workspace/apps-web/blog-v2
test: make check
protected: ["*.tsp", "*_test.go"]
dependencies: [go.mod, go.sum]
---

## 📄 feat: feed (feed)

- [ ] add it
  - DONE go test ./... passes
`

func TestParseMetaReadsTheFrontmatter(t *testing.T) {
	home, _ := os.UserHomeDir()
	meta, err := ParseMeta(withFrontmatter)
	want := Meta{
		Repo:         filepath.Join(home, "workspace/apps-web/blog-v2"),
		Test:         "make check",
		Protected:    []string{"*.tsp", "*_test.go"},
		Dependencies: []string{"go.mod", "go.sum"},
	}
	if err != nil || !reflect.DeepEqual(meta, want) {
		t.Fatalf("got %+v, %v; want %+v", meta, err, want)
	}
	if sections := Parse(withFrontmatter); len(sections) != 1 || sections[0].Slug != "feed" {
		t.Fatalf("the frontmatter should leave the headings as they were, got %+v", sections)
	}
}

func TestParseMetaWithoutFrontmatterIsEmpty(t *testing.T) {
	if meta, err := ParseMeta(oneSection()); err != nil || !reflect.DeepEqual(meta, Meta{}) {
		t.Fatalf("got %+v, %v; want an empty Meta", meta, err)
	}
}

func TestParseMetaNotClosedFails(t *testing.T) {
	if _, err := ParseMeta("---\nrepo: x\n\n## 📄 feat: feed (feed)\n"); err == nil || err.Error() != "plan: frontmatter not closed" {
		t.Fatalf("got %v, want plan: frontmatter not closed", err)
	}
}
