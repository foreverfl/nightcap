package claude

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// RunRound runs one `claude -p` round inside dir and writes its stdout and stderr to logPath.
// It returns an error when claude cannot start, the log cannot be written, or the output is not a result JSON.
func RunRound(ctx context.Context, dir, prompt string, maxTurns int, allowedTools []string, logPath string) (Result, error) {
	command := exec.CommandContext(ctx, claudeBinary(),
		"-p", prompt,
		"--max-turns", strconv.Itoa(maxTurns),
		"--output-format", "json",
		"--allowedTools", strings.Join(allowedTools, ","),
	)
	command.Dir = dir
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runErr := command.Run()

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return Result{}, err
	}
	logText := stdout.String() + "\n--- stderr ---\n" + stderr.String()
	if err := os.WriteFile(logPath, []byte(logText), 0o644); err != nil {
		return Result{}, err
	}

	result, parseErr := ParseResult(stdout.Bytes())
	if parseErr != nil && runErr != nil {
		return Result{}, fmt.Errorf("claude: %w (log: %s)", runErr, logPath)
	}
	return result, parseErr
}
