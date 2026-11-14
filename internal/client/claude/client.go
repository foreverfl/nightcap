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

	"nightcap/internal/indent"
)

// RunRound runs one `claude -p` round in worktreeDir, appending each output line to logPath and what claude does to screen as it comes; a non-empty sessionID resumes that conversation.
// When authorName and authorEmail are both set, the commits claude makes carry them as author and committer.
// It fails when claude cannot start, the log cannot be written, or the last line is not a result JSON.
func RunRound(ctx context.Context, worktreeDir, prompt, sessionID string, maxTurns int, allowedTools []string, authorName, authorEmail, logPath string, screen io.Writer) (Result, error) {
	command := exec.CommandContext(ctx, claudeBinary(),
		"-p", prompt,
		"--model", "opus",
		"--max-turns", strconv.Itoa(maxTurns),
		"--output-format", "stream-json", "--verbose",
		"--allowedTools", strings.Join(allowedTools, ","),
	)
	if sessionID != "" {
		command.Args = append(command.Args, "--resume", sessionID)
	}
	command.Dir = worktreeDir
	if authorName != "" && authorEmail != "" {
		command.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME="+authorName, "GIT_AUTHOR_EMAIL="+authorEmail,
			"GIT_COMMITTER_NAME="+authorName, "GIT_COMMITTER_EMAIL="+authorEmail,
		)
	}
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
	lastLine, copyErr := copyLinesToLogAndScreen(stdoutPipe, logFile, screen, worktreeDir)
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

// copyLinesToLogAndScreen writes each line of claudeOutput to logFile as soon as it arrives and its ScreenLines to screen, indented,
// and returns the last line: the result JSON in stream-json.
func copyLinesToLogAndScreen(claudeOutput io.Reader, logFile, screen io.Writer, worktreeDir string) ([]byte, error) {
	scanner := bufio.NewScanner(claudeOutput)
	scanner.Buffer(nil, 1024*1024) // one stream-json line holds a whole file claude read
	var lastLine []byte
	for scanner.Scan() {
		if _, err := fmt.Fprintf(logFile, "%s\n", scanner.Bytes()); err != nil {
			return nil, err
		}
		for _, screenLine := range ScreenLines(scanner.Bytes(), worktreeDir) {
			indent.Fprintf(screen, 2, "%s\n", screenLine)
		}
		lastLine = bytes.Clone(scanner.Bytes())
	}
	return lastLine, scanner.Err()
}
