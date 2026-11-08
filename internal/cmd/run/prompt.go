package run

import (
	_ "embed"
	"strings"
)

//go:embed prompts/commit.md
var commitRules string

// BuildPrompt returns one round's prompt: the commit rules and commitBody, then failureTail from the second round on.
// When isResumed it leaves out what the session already has: the rules, and on a retry commitBody too.
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
