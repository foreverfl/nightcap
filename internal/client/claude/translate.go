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

// Translate asks claude to turn the Korean PR body into English and returns the translation.
func Translate(ctx context.Context, korean string) (string, error) {
	return askOnce(ctx, "translate", translateInstructions+"\n\n"+korean)
}

var translateTitleInstructions = "Translate this PR title from Korean to English as one line like a commit subject, starting with a lowercase verb. Reply with the translation only."

// TranslateTitle asks claude to turn the Korean PR title, without its type, into one English line, e.g. "show the day's real ranking on the first screen".
func TranslateTitle(ctx context.Context, korean string) (string, error) {
	english, err := askOnce(ctx, "translate title", translateTitleInstructions+"\n\n"+korean)
	return strings.TrimSpace(english), err
}
