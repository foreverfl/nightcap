package run

import (
	_ "embed"
	"strings"
)

//go:embed prompts/unit.md
var unitRules string

// BuildPrompt joins the session marker, the unit rules, the unit's plan text and,
// from the second round on, the tail of the previous failing check.
func BuildPrompt(unitBody, failureTail string) string {
	parts := []string{"This session is a nightcap run.", unitRules, "## Unit\n\n" + unitBody}
	if failureTail != "" {
		parts = append(parts, "## The check failed last round\n\n```\n"+failureTail+"\n```")
	}
	return strings.Join(parts, "\n\n")
}
