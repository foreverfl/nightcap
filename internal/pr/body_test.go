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
