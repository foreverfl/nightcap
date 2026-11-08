package test

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

const tailLines = 40

// Result is one test run: Green when the command exited 0, Tail the last lines of its output.
type Result struct {
	Green bool
	Tail  string
}

// Run runs command with `sh -c` in worktreeDir and returns whether it exited 0, with the last 40 lines of its output.
// A non-zero exit is a red Result, not an error; it fails only when sh cannot start.
func Run(ctx context.Context, worktreeDir, command string) (Result, error) {
	shell := exec.CommandContext(ctx, "sh", "-c", command)
	shell.Dir = worktreeDir
	output, runErr := shell.CombinedOutput()
	result := Result{Green: runErr == nil, Tail: tail(string(output))}
	var exitErr *exec.ExitError
	if runErr != nil && !errors.As(runErr, &exitErr) {
		return result, runErr
	}
	return result, nil
}

// tail returns the last 40 lines of text, or all of it when it is shorter.
func tail(text string) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) > tailLines {
		lines = lines[len(lines)-tailLines:]
	}
	return strings.Join(lines, "\n")
}
