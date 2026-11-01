package plan

import (
	"strings"
	"testing"
)

// oneSection is the smallest plan that yields one section with one open unit.
func oneSection() string {
	return "## feat: 홈 화면 피드 (home-feed)\n" +
		"\n" +
		"Show the feed on the home screen.\n" +
		"\n" +
		"- [ ] add the list endpoint\n" +
		"  - WHAT  return posts\n" +
		"  - DONE  go test ./internal/feed passes\n"
}

func TestParseReadsHeading(t *testing.T) {
	section := Parse(oneSection())[0]
	if section.Type != "feat" || section.Title != "홈 화면 피드" || section.Slug != "home-feed" {
		t.Fatalf("got %q %q %q", section.Type, section.Title, section.Slug)
	}
}

func TestParseReadsSummary(t *testing.T) {
	section := Parse(oneSection())[0]
	if section.Summary != "Show the feed on the home screen." {
		t.Fatalf("Summary = %q", section.Summary)
	}
}

func TestParseReadsUnitBody(t *testing.T) {
	unit := Parse(oneSection())[0].Units[0]
	want := "- [ ] add the list endpoint\n  - WHAT  return posts\n  - DONE  go test ./internal/feed passes"
	if unit.Title != "add the list endpoint" || unit.Body != want {
		t.Fatalf("got %q / %q", unit.Title, unit.Body)
	}
}

func TestParseSkipsDoneUnits(t *testing.T) {
	text := strings.Replace(oneSection(), "- [ ]", "- [x]", 1)
	if units := Parse(text)[0].Units; len(units) != 0 {
		t.Fatalf("got %d units, want 0", len(units))
	}
}

func TestParseIgnoresCheckboxInsideFence(t *testing.T) {
	text := oneSection() + "\n```text\n- [ ] not a unit\n```\n"
	if units := Parse(text)[0].Units; len(units) != 1 {
		t.Fatalf("got %d units, want 1", len(units))
	}
}

func TestParseMarksUnitWithoutCheckName(t *testing.T) {
	text := strings.Replace(oneSection(), "go test ./internal/feed passes", "read it by hand", 1)
	if Parse(text)[0].Units[0].HasCheck {
		t.Fatal("HasCheck = true, want false")
	}
}

func TestParseSkipsHeadingWithoutType(t *testing.T) {
	text := strings.Replace(oneSection(), "feat: 홈 화면 피드 (home-feed)", "홈 화면 피드", 1)
	if sections := Parse(text); len(sections) != 0 {
		t.Fatalf("got %d sections, want 0", len(sections))
	}
}

func TestPickSkipsSectionWithoutCheckableUnit(t *testing.T) {
	manual := strings.Replace(oneSection(), "go test ./internal/feed passes", "read it by hand", 1)
	second := strings.Replace(oneSection(), "(home-feed)", "(home-feed-2)", 1)
	section, ok := Pick(Parse(manual+"\n"+second), "")
	if !ok || section.Slug != "home-feed-2" {
		t.Fatalf("got %q, %v; want home-feed-2", section.Slug, ok)
	}
}

func TestPickByName(t *testing.T) {
	second := strings.Replace(oneSection(), "(home-feed)", "(home-feed-2)", 1)
	section, ok := Pick(Parse(oneSection()+"\n"+second), "home-feed-2")
	if !ok || section.Slug != "home-feed-2" {
		t.Fatalf("got %q, %v; want home-feed-2", section.Slug, ok)
	}
}
