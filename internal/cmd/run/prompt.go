package run

import (
	_ "embed"
	"strings"
)

//go:embed prompts/commit.md
var commitRules string

// BuildPrompt joins the session marker, the commit rules, the commit's plan text and,
// from the second round on, the tail of the previous failing test.
// A resumed session already has the marker and the rules, and on a retry the commit too, so those are left out.
func BuildPrompt(commitBody, failureTail string, isResumed bool) string {
	var parts []string
	if !isResumed {
		parts = append(parts, "This session is a nightcap run.", commitRules)
	}
	if !isResumed || failureTail == "" {
		parts = append(parts, "## Commit\n\n"+commitBody)
	}
	if failureTail != "" {
		parts = append(parts, "## The test failed last round\n\n```\n"+failureTail+"\n```")
	}
	return strings.Join(parts, "\n\n")
}
