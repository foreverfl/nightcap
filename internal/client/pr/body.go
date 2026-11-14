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

	IsMissingEnglish     bool // translation failed, so the risk section opens with "no English block"
	IsMissingReviewOrder bool // claude could not write the review order, so the risk section says "no review order"
}

// HasNoRisk reports whether the risk section would be empty: no missing English or review order, no read-code reason and no risk line.
func (input BodyInput) HasNoRisk() bool {
	hasReadCode := slices.ContainsFunc(input.Commits, func(commit BodyCommit) bool { return commit.ReadCodeReason != "" })
	return !input.IsMissingEnglish && !input.IsMissingReviewOrder && !hasReadCode && len(input.Risks) == 0
}

// BuildBody fills prompts/pr.md with input and returns the Korean PR body.
func BuildBody(input BodyInput) (string, error) {
	var body strings.Builder
	if err := bodyTemplate.Execute(&body, input); err != nil {
		return "", err
	}
	return body.String(), nil
}

// BuildBilingualBody returns the English body translated from the Korean one, a "---", then the Korean one.
// When translate fails it returns the Korean one alone, with "no English block" first in its risks.
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
