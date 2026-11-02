package check

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

const tailLines = 40

// Result is one check run: Green when the command exited 0, Tail the last lines of its output.
type Result struct {
	Green bool
	Tail  string
}

// Run runs command with `sh -c` inside dir and returns Green when it exited 0,
// with the last 40 lines of stdout and stderr as Tail. A non-zero exit is a red Result, not an error;
// it returns an error only when sh itself cannot start.
func Run(ctx context.Context, dir, command string) (Result, error) {
	shell := exec.CommandContext(ctx, "sh", "-c", command)
	shell.Dir = dir
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
