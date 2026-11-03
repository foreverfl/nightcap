package pr

import (
	_ "embed"
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

// BodyInput is everything the PR description is built from: the line under the heading, the finished units and the last check's tail.
type BodyInput struct {
	Summary   string
	Units     []BodyUnit
	CheckTail string
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
