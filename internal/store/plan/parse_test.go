package plan

import (
	"os"
	"strings"
	"testing"
)

// onePR is the smallest plan that yields one PR with one open commit.
func onePR() string {
	return "## 📄 feat: 홈 화면 피드 (home-feed)\n" +
		"\n" +
		"Show the feed on the home screen.\n" +
		"\n" +
		"- [ ] add the list endpoint\n" +
		"  - WHAT  return posts\n" +
		"  - DONE  go test ./internal/feed passes\n"
}

func TestParseReadsPR(t *testing.T) {
	planPR := Parse(onePR())[0]
	if planPR.Type != "feat" || planPR.Title != "홈 화면 피드" || planPR.Slug != "home-feed" {
		t.Fatalf("got %q %q %q", planPR.Type, planPR.Title, planPR.Slug)
	}
}

func TestParseReadsSummary(t *testing.T) {
	planPR := Parse(onePR())[0]
	if planPR.Summary != "Show the feed on the home screen." {
		t.Fatalf("Summary = %q", planPR.Summary)
	}
}

func TestParseReadsCommitBody(t *testing.T) {
	commit := Parse(onePR())[0].Commits[0]
	want := "- [ ] add the list endpoint\n  - WHAT  return posts\n  - DONE  go test ./internal/feed passes"
	if commit.Title != "add the list endpoint" || commit.Body != want {
		t.Fatalf("got %q / %q", commit.Title, commit.Body)
	}
}

func TestParseSkipsDoneCommits(t *testing.T) {
	text := strings.Replace(onePR(), "- [ ]", "- [x]", 1)
	if commits := Parse(text)[0].Commits; len(commits) != 0 {
		t.Fatalf("got %d commits, want 0", len(commits))
	}
}

func TestParseIgnoresCheckboxInsideFence(t *testing.T) {
	text := onePR() + "\n```text\n- [ ] not a commit\n```\n"
	if commits := Parse(text)[0].Commits; len(commits) != 1 {
		t.Fatalf("got %d commits, want 1", len(commits))
	}
}

func TestParseMarksCommitWithoutTestName(t *testing.T) {
	text := strings.Replace(onePR(), "go test ./internal/feed passes", "read it by hand", 1)
	if Parse(text)[0].Commits[0].HasTest {
		t.Fatal("HasTest = true, want false")
	}
}

func TestParseSkipsPRWithoutType(t *testing.T) {
	text := strings.Replace(onePR(), "feat: 홈 화면 피드 (home-feed)", "홈 화면 피드", 1)
	if planPRs := Parse(text); len(planPRs) != 0 {
		t.Fatalf("got %d planPRs, want 0", len(planPRs))
	}
}

func TestPickSkipsPRWithoutTestableCommit(t *testing.T) {
	manual := strings.Replace(onePR(), "go test ./internal/feed passes", "read it by hand", 1)
	second := strings.Replace(onePR(), "(home-feed)", "(home-feed-2)", 1)
	planPR, ok := Pick(Parse(manual + "\n" + second))
	if !ok || planPR.Slug != "home-feed-2" {
		t.Fatalf("got %q, %v; want home-feed-2", planPR.Slug, ok)
	}
}

func TestParseReadsExamplePlans(t *testing.T) {
	for _, path := range []string{"../../../examples/plan.md", "../../../examples/plan-ko.md"} {
		text, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		planPRs := Parse(string(text))
		if len(planPRs) != 1 {
			t.Fatalf("%s: got %d planPRs, want 1", path, len(planPRs))
		}
		withTest := 0
		for _, commit := range planPRs[0].Commits {
			if commit.HasTest {
				withTest++
			}
		}
		if len(planPRs[0].Commits) != 3 || withTest != 2 {
			t.Fatalf("%s: got %d commits, %d with a test; want 3 and 2", path, len(planPRs[0].Commits), withTest)
		}
	}
}

func TestPickNeverTakesAPersonPR(t *testing.T) {
	person := strings.Replace(onePR(), "## 📄 feat:", "## 👤 feat:", 1)
	planPRs := Parse(person)
	if len(planPRs) != 1 || !planPRs[0].IsPerson() || planPRs[0].Type != "feat" {
		t.Fatalf("a 👤 PR should parse as a person PR of type feat, got %+v", planPRs)
	}
	if _, ok := Pick(planPRs); ok {
		t.Fatal("Pick should not take a 👤 PR")
	}
}

func TestParseReadsEachReadingMark(t *testing.T) {
	want := map[string]Mark{"🤖": MarkAuto, "📄": MarkReadPR, "👤": MarkPerson}
	for symbol, mark := range want {
		planPR := Parse(strings.Replace(onePR(), "📄", symbol, 1))[0]
		if planPR.Mark != mark || planPR.Type != "feat" {
			t.Fatalf("%s: got mark %q type %q, want %q feat", symbol, planPR.Mark, planPR.Type, mark)
		}
	}
}

func TestPickSkipsAPRWithoutAMark(t *testing.T) {
	planPRs := Parse(strings.Replace(onePR(), "📄 ", "", 1))
	if len(planPRs) != 1 || planPRs[0].Mark != MarkNone {
		t.Fatalf("an unmarked PR should still parse with MarkNone, got %+v", planPRs)
	}
	if _, ok := Pick(planPRs); ok {
		t.Fatal("Pick should not take an unmarked PR")
	}
}
