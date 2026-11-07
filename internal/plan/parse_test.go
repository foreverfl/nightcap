package plan

import (
	"os"
	"strings"
	"testing"
)

// oneSection is the smallest plan that yields one section with one open unit.
func oneSection() string {
	return "## 📄 feat: 홈 화면 피드 (home-feed)\n" +
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

func TestParseReadsExamplePlans(t *testing.T) {
	for _, path := range []string{"../../examples/plan.md", "../../examples/plan-ko.md"} {
		text, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sections := Parse(string(text))
		if len(sections) != 1 {
			t.Fatalf("%s: got %d sections, want 1", path, len(sections))
		}
		checked := 0
		for _, unit := range sections[0].Units {
			if unit.HasCheck {
				checked++
			}
		}
		if len(sections[0].Units) != 3 || checked != 2 {
			t.Fatalf("%s: got %d units, %d with a check; want 3 and 2", path, len(sections[0].Units), checked)
		}
	}
}

func TestPickNeverTakesAPersonHeading(t *testing.T) {
	person := strings.Replace(oneSection(), "## 📄 feat:", "## 👤 feat:", 1)
	sections := Parse(person)
	if len(sections) != 1 || !sections[0].IsPerson() || sections[0].Type != "feat" {
		t.Fatalf("a 👤 heading should parse as a person section of type feat, got %+v", sections)
	}
	if _, ok := Pick(sections, ""); ok {
		t.Fatal("Pick should not take a 👤 heading")
	}
	if _, ok := Pick(sections, "home-feed"); ok {
		t.Fatal("Pick should not take a 👤 heading even by name")
	}
}

func TestParseReadsEachReadingMark(t *testing.T) {
	want := map[string]Mark{"🤖": MarkAuto, "📄": MarkReadPR, "👀": MarkReadCode, "👤": MarkPerson}
	for symbol, mark := range want {
		section := Parse(strings.Replace(oneSection(), "📄", symbol, 1))[0]
		if section.Mark != mark || section.Type != "feat" {
			t.Fatalf("%s: got mark %q type %q, want %q feat", symbol, section.Mark, section.Type, mark)
		}
	}
}

func TestPickSkipsAHeadingWithoutAMark(t *testing.T) {
	sections := Parse(strings.Replace(oneSection(), "📄 ", "", 1))
	if len(sections) != 1 || sections[0].Mark != MarkNone {
		t.Fatalf("an unmarked heading should still parse with MarkNone, got %+v", sections)
	}
	if _, ok := Pick(sections, ""); ok {
		t.Fatal("Pick should not take an unmarked heading")
	}
}
