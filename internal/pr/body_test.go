package pr

import (
	"errors"
	"strings"
	"testing"
)

func TestBuildBodyListsUnitsAndTheReadCodeReason(t *testing.T) {
	body, err := BuildBody(BodyInput{
		Summary: "피드를 보여 준다",
		Commits: []BodyCommit{
			{What: "피드 목록", Done: "go test ./internal/feed 통과"},
			{What: "피드 화면", Done: "go test . 통과", ReadCodeReason: "outside WHERE: main.go"},
		},
		TestTail: "ok  nightcap 0.9s",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		"## What",
		"피드를 보여 준다",
		"",
		"## Review order",
		"- 피드 목록: go test ./internal/feed 통과",
		"- 피드 화면: go test . 통과",
		"",
		"## Risks",
		"- outside WHERE: main.go",
		"",
		"## Verified",
		"```",
		"ok  nightcap 0.9s",
		"```",
	}, "\n") + "\n"
	if body != want {
		t.Fatalf("got:\n%s\nwant:\n%s", body, want)
	}
}

// riskSection returns the lines of body between "## Risks" and "## Verified".
func riskSection(body string) string {
	_, afterRisk, _ := strings.Cut(body, "## Risks\n")
	risk, _, _ := strings.Cut(afterRisk, "\n## Verified")
	return risk
}

func TestBuildBodyListsRiskLines(t *testing.T) {
	body, err := BuildBody(BodyInput{Risks: []string{"env: internal/feed/client.go"}})
	if err != nil {
		t.Fatal(err)
	}
	if risk := riskSection(body); risk != "- env: internal/feed/client.go\n" {
		t.Fatalf("the risk section should hold the one risk line, got %q", risk)
	}
}

func TestBuildBodyWithNothingRiskySaysNone(t *testing.T) {
	body, err := BuildBody(BodyInput{})
	if err != nil {
		t.Fatal(err)
	}
	if risk := riskSection(body); risk != "- none\n" {
		t.Fatalf("with nothing risky the risk section should say none, got %q", risk)
	}
}

func TestBuildBilingualBodyWithoutTranslationSaysEnglishIsMissing(t *testing.T) {
	failingTranslate := func(korean string) (string, error) { return "", errors.New("claude down") }
	body, err := BuildBilingualBody(BodyInput{}, failingTranslate)
	if err != nil {
		t.Fatal(err)
	}
	if risk := riskSection(body); risk != "- no English block\n" {
		t.Fatalf("a failed translation should leave only no English block in the risk section, got %q", risk)
	}
}

func TestBuildBodyPutsTheReviewOrderInPlaceOfTheUnitList(t *testing.T) {
	body, err := BuildBody(BodyInput{
		Commits:     []BodyCommit{{What: "피드 목록", Done: "go test 통과"}},
		ReviewOrder: "흐름 1. 실행할 때: 피드를 읽는다",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "## Review order\n흐름 1. 실행할 때: 피드를 읽는다\n") {
		t.Fatalf("the review order should sit under ## Review order, got:\n%s", body)
	}
	if strings.Contains(body, "- 피드 목록: go test 통과") {
		t.Fatalf("with a review order the commit list should be gone, got:\n%s", body)
	}
}

func TestBuildBodyWhenClaudeFailedTheReviewOrderWarnsInTheRiskSection(t *testing.T) {
	body, err := BuildBody(BodyInput{IsMissingReviewOrder: true})
	if err != nil {
		t.Fatal(err)
	}
	if risk := riskSection(body); risk != "- no review order\n" {
		t.Fatalf("a failed review order should leave only no review order in the risk section, got %q", risk)
	}
}
