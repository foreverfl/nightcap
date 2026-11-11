package claude

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// RunRound runs one `claude -p` round in worktreeDir, appending each output line to logPath as it comes; a non-empty sessionID resumes that conversation.
// It fails when claude cannot start, the log cannot be written, or the last line is not a result JSON.
func RunRound(ctx context.Context, worktreeDir, prompt, sessionID string, maxTurns int, allowedTools []string, logPath string) (Result, error) {
	command := exec.CommandContext(ctx, claudeBinary(),
		"-p", prompt,
		"--max-turns", strconv.Itoa(maxTurns),
		"--output-format", "stream-json", "--verbose",
		"--allowedTools", strings.Join(allowedTools, ","),
	)
	if sessionID != "" {
		command.Args = append(command.Args, "--resume", sessionID)
	}
	command.Dir = worktreeDir
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return Result{}, err
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return Result{}, err
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return Result{}, err
	}
	defer logFile.Close()

	if err := command.Start(); err != nil {
		return Result{}, fmt.Errorf("claude: %w", err)
	}
	lastLine, copyErr := copyLinesToLog(stdoutPipe, logFile)
	if copyErr != nil {
		io.Copy(io.Discard, stdoutPipe) // drain, or claude blocks on a full pipe and Wait never returns
	}
	runErr := command.Wait()
	if _, err := logFile.WriteString("\n--- stderr ---\n" + stderr.String()); err != nil {
		return Result{}, err
	}
	if copyErr != nil {
		return Result{}, copyErr
	}

	result, parseErr := ParseResult(lastLine)
	if parseErr != nil && runErr != nil {
		return Result{}, fmt.Errorf("claude: %w (log: %s)", runErr, logPath)
	}
	return result, parseErr
}

// copyLinesToLog writes each line of claudeOutput to logFile as soon as it arrives, and returns the last line: the result JSON in stream-json.
func copyLinesToLog(claudeOutput io.Reader, logFile io.Writer) ([]byte, error) {
	scanner := bufio.NewScanner(claudeOutput)
	scanner.Buffer(nil, 1024*1024) // one stream-json line holds a whole file claude read
	var lastLine []byte
	for scanner.Scan() {
		if _, err := fmt.Fprintf(logFile, "%s\n", scanner.Bytes()); err != nil {
			return nil, err
		}
		lastLine = bytes.Clone(scanner.Bytes())
	}
	return lastLine, scanner.Err()
}
