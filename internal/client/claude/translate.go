package claude

import (
	"context"
	"strings"
)

var translateInstructions = strings.Join([]string{
	"Translate this GitHub PR description from Korean to English.",
	"Headings: 무엇을 → What, 읽는 순서 → Review order, 위험 → Risks, 확인한 것 → Verified.",
	"Keep the markdown, code, paths and identifiers as they are. Reply with the translation only.",
}, "\n")

// Translate asks claude, in one turn with no tools, to turn the Korean PR description korean into English,
// and returns the translation. It returns an error when claude fails or does not answer with a finished result.
func Translate(ctx context.Context, korean string) (string, error) {
	return askOnce(ctx, "translate", translateInstructions+"\n\n"+korean)
}
