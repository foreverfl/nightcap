package pr

import (
	"slices"
	"strings"

	"nightcap/internal/client/git"
)

// riskWords maps what a reviewer should look at to the words that mark it on an added line.
var riskWords = []struct {
	label string
	words []string
}{
	{"환경 변수 읽음", []string{"os.Getenv", "env::var", "process.env"}},
	{"네트워크 호출", []string{"http.Get", "http.Post", "http.NewRequest", "net.Dial", "reqwest::", "fetch("}},
	{"삭제", []string{"os.Remove", "DELETE FROM", "DROP TABLE", "rm -rf"}},
}

var signatureStarts = []string{"func ", "pub fn "}

var routeMarks = []string{".route(", "@route"}

// RiskLines returns one line per thing in the branch's diff a reviewer should look at: files whose added lines
// read env, call the network or delete, added function signatures and routes, and changed .sql files. Empty when none.
func RiskLines(changes []git.Change, addedLines []git.AddedLine) []string {
	var risks []string
	for _, line := range addedLines {
		text := strings.TrimSpace(line.Text)
		for _, risk := range riskWords {
			fileRisk := risk.label + ": " + line.Path
			if containsAny(text, risk.words) && !slices.Contains(risks, fileRisk) {
				risks = append(risks, fileRisk)
			}
		}
		if slices.ContainsFunc(signatureStarts, func(start string) bool { return strings.HasPrefix(text, start) }) {
			risks = append(risks, "새 함수: "+line.Path+": "+text)
		}
		if containsAny(text, routeMarks) {
			risks = append(risks, "새 라우트: "+line.Path+": "+text)
		}
	}
	for _, change := range changes {
		if strings.HasSuffix(change.Path, ".sql") {
			risks = append(risks, "SQL 파일: "+change.Path)
		}
	}
	return risks
}

// containsAny reports whether text contains one of words.
func containsAny(text string, words []string) bool {
	return slices.ContainsFunc(words, func(word string) bool { return strings.Contains(text, word) })
}
