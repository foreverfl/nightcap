package pr

import (
	_ "embed"
	"slices"
	"strings"
	"text/template"
)

//go:embed prompts/pr.md
var bodyTemplateText string

var bodyTemplate = template.Must(template.New("pr").Parse(bodyTemplateText))

// BodyCommit is one finished commit as the PR description lists it; ReadCodeReason is empty unless its verdict was read-code.
type BodyCommit struct {
	What           string
	Done           string
	ReadCodeReason string
}

// BodyInput is everything the PR description is built from: the line under the ## line, the finished commits,
// the lines RiskLines picked from the diff, the flows ReviewOrder wrote and the last check's tail.
type BodyInput struct {
	Summary     string
	Commits     []BodyCommit
	Risks       []string
	TestTail    string
	ReviewOrder string // the flows claude wrote from the diff; empty means list each commit's WHAT and DONE instead

	IsMissingEnglish     bool // translation failed, so the risk section opens with "영어 블록 없음(번역 실패)"
	IsMissingReviewOrder bool // claude could not write the review order, so the risk section says "읽는 순서 없음"
}

// HasNoRisk reports whether the risk section would be empty: no missing English or review order, no read-code reason and no risk line.
func (input BodyInput) HasNoRisk() bool {
	hasReadCode := slices.ContainsFunc(input.Commits, func(commit BodyCommit) bool { return commit.ReadCodeReason != "" })
	return !input.IsMissingEnglish && !input.IsMissingReviewOrder && !hasReadCode && len(input.Risks) == 0
}

// UniqueRisks returns the commits' read-code reasons, then the diff's risk lines, each once, in the order they first appear.
func (input BodyInput) UniqueRisks() []string {
	var candidates []string
	for _, commit := range input.Commits {
		if commit.ReadCodeReason != "" {
			candidates = append(candidates, commit.ReadCodeReason)
		}
	}
	candidates = append(candidates, input.Risks...)
	var unique []string
	for _, line := range candidates {
		if !slices.Contains(unique, line) {
			unique = append(unique, line)
		}
	}
	return unique
}

// koreanReasonStarts maps the start of a reason verdict.Classify gives to the Korean the PR body shows.
var koreanReasonStarts = []struct{ english, korean string }{
	{"outside WHERE: ", "WHERE 밖의 파일: "},
	{"added a migration: ", "새 마이그레이션: "},
}

// koreanReason returns reason with its start in Korean, e.g. "outside WHERE: main.go" → "WHERE 밖의 파일: main.go"; a start not in the table stays as it is.
func koreanReason(reason string) string {
	for _, start := range koreanReasonStarts {
		if rest, isFound := strings.CutPrefix(reason, start.english); isFound {
			return start.korean + rest
		}
	}
	return reason
}

// BuildBody fills prompts/pr.md with input and returns the Korean PR body.
func BuildBody(input BodyInput) (string, error) {
	// a copy, so the caller's commits keep the English reason
	input.Commits = slices.Clone(input.Commits)
	for i := range input.Commits {
		input.Commits[i].ReadCodeReason = koreanReason(input.Commits[i].ReadCodeReason)
	}
	var body strings.Builder
	if err := bodyTemplate.Execute(&body, input); err != nil {
		return "", err
	}
	return body.String(), nil
}

// BuildBilingualBody returns the English body translated from the Korean one, a "---", then the Korean one.
// When translate fails it returns the Korean one alone, with "영어 블록 없음(번역 실패)" first in its risks.
func BuildBilingualBody(input BodyInput, translate func(korean string) (string, error)) (string, error) {
	korean, err := BuildBody(input)
	if err != nil {
		return "", err
	}
	english, translateErr := translate(korean)
	if translateErr != nil {
		input.IsMissingEnglish = true
		return BuildBody(input)
	}
	return english + "\n\n---\n\n" + korean, nil
}
