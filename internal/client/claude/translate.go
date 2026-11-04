package claude

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
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

// askOnce runs claude on prompt in one turn with no tools and returns its answer; what names the call in errors.
// It returns an error when claude fails or does not answer with a finished result.
func askOnce(ctx context.Context, what, prompt string) (string, error) {
	command := exec.CommandContext(ctx, claudeBinary(), "-p", prompt, "--max-turns", "1", "--output-format", "json")
	output, err := command.Output()
	if exitErr, isExitErr := errors.AsType[*exec.ExitError](err); isExitErr {
		return "", fmt.Errorf("claude %s: %w\n%s", what, err, exitErr.Stderr)
	}
	if err != nil {
		return "", err
	}
	result, err := ParseResult(output)
	if err != nil {
		return "", err
	}
	if result.Outcome != Done {
		return "", fmt.Errorf("claude %s ended %s: %s", what, result.Outcome, result.Text)
	}
	return result.Text, nil
}
