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

// BodyUnit is one finished unit as the PR description lists it; ReadCodeReason is empty unless its verdict was read-code.
type BodyUnit struct {
	What           string
	Done           string
	ReadCodeReason string
}

// BodyInput is everything the PR description is built from: the line under the heading, the finished units,
// the lines RiskLines picked from the diff, the flows ReviewOrder wrote and the last check's tail.
type BodyInput struct {
	Summary     string
	Units       []BodyUnit
	Risks       []string
	CheckTail   string
	ReviewOrder string // the flows claude wrote from the diff; empty means list each unit's WHAT and DONE instead

	IsMissingEnglish     bool // translation failed, so the risk section opens with "영문 없음"
	IsMissingReviewOrder bool // claude could not write the review order, so the risk section says "읽는 순서 없음"
}

// HasNoRisk reports whether the risk section would be empty: no missing English or review order, no read-code reason and no risk line.
func (input BodyInput) HasNoRisk() bool {
	hasReadCode := slices.ContainsFunc(input.Units, func(unit BodyUnit) bool { return unit.ReadCodeReason != "" })
	return !input.IsMissingEnglish && !input.IsMissingReviewOrder && !hasReadCode && len(input.Risks) == 0
}

// BuildBody fills prompts/pr.md with input and returns the Korean half of the PR description.
// It returns an error only when the template cannot be filled.
func BuildBody(input BodyInput) (string, error) {
	var body strings.Builder
	if err := bodyTemplate.Execute(&body, input); err != nil {
		return "", err
	}
	return body.String(), nil
}

// BuildBilingualBody returns the English description translated from the Korean one, a "---", then the Korean one.
// When translate fails it returns the Korean one alone with "영문 없음" first in the risk section;
// it returns an error only when the template cannot be filled.
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
