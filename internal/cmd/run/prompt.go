package run

import (
	_ "embed"
	"strings"
)

//go:embed prompts/unit.md
var commitRules string

// BuildPrompt joins the session marker, the unit rules, the unit's plan text and,
// from the second round on, the tail of the previous failing check.
// A resumed session already has the marker and the rules, and on a retry the unit too, so those are left out.
func BuildPrompt(commitBody, failureTail string, isResumed bool) string {
	var parts []string
	if !isResumed {
		parts = append(parts, "This session is a nightcap run.", commitRules)
	}
	if !isResumed || failureTail == "" {
		parts = append(parts, "## Unit\n\n"+commitBody)
	}
	if failureTail != "" {
		parts = append(parts, "## The check failed last round\n\n```\n"+failureTail+"\n```")
	}
	return strings.Join(parts, "\n\n")
}
