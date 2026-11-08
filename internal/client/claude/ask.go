package claude

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// askOnce runs claude on prompt in one turn with no tools and returns its answer; callName names the call in errors.
func askOnce(ctx context.Context, callName, prompt string) (string, error) {
	command := exec.CommandContext(ctx, claudeBinary(), "-p", prompt, "--max-turns", "1", "--output-format", "json")
	output, err := command.Output()
	if exitErr, isExitErr := errors.AsType[*exec.ExitError](err); isExitErr {
		return "", fmt.Errorf("claude %s: %w\n%s", callName, err, exitErr.Stderr)
	}
	if err != nil {
		return "", err
	}
	result, err := ParseResult(output)
	if err != nil {
		return "", err
	}
	if result.Outcome != Done {
		return "", fmt.Errorf("claude %s ended %s: %s", callName, result.Outcome, result.Text)
	}
	return result.Text, nil
}
