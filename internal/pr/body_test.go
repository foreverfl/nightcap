package pr

import (
	"strings"
	"testing"
)

func TestBuildBodyListsUnitsAndTheReadCodeReason(t *testing.T) {
	body, err := BuildBody(BodyInput{
		Summary: "피드를 보여 준다",
		Units: []BodyUnit{
			{What: "피드 목록", Done: "go test ./internal/feed 통과"},
			{What: "피드 화면", Done: "go test . 통과", ReadCodeReason: "outside WHERE: main.go"},
		},
		CheckTail: "ok  nightcap 0.9s",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		"## 무엇을",
		"피드를 보여 준다",
		"",
		"## 읽는 순서",
		"- 피드 목록: go test ./internal/feed 통과",
		"- 피드 화면: go test . 통과",
		"",
		"## 위험",
		"- 코드까지 읽어야 함: outside WHERE: main.go",
		"",
		"## 확인한 것",
		"```",
		"ok  nightcap 0.9s",
		"```",
	}, "\n") + "\n"
	if body != want {
		t.Fatalf("got:\n%s\nwant:\n%s", body, want)
	}
}

// riskSection returns the lines of body between "## 위험" and "## 확인한 것".
func riskSection(body string) string {
	_, afterRisk, _ := strings.Cut(body, "## 위험\n")
	risk, _, _ := strings.Cut(afterRisk, "\n## 확인한 것")
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
	if risk := riskSection(body); risk != "- 없음\n" {
		t.Fatalf("with nothing risky the risk section should say 없음, got %q", risk)
	}
}
