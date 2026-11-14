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

// Round is what one `claude -p` round needs.
type Round struct {
	WorktreeDir  string   // where claude runs and edits files
	Prompt       string   // the instruction sent with -p
	SessionID    string   // when not empty, the round resumes that conversation
	MaxTurns     int      // passed as --max-turns
	AllowedTools []string // passed as --allowedTools, joined with commas
	AuthorName   string   // with AuthorEmail, the author and committer of the commits claude makes
	AuthorEmail  string   // with AuthorName, the author and committer of the commits claude makes
	LogPath      string   // where every output line of claude is written
}

// RunRound runs one claude round, writing each output line to round.LogPath and what claude does to screen as it comes.
// It fails when claude cannot start, the log cannot be written, or the last line is not a result JSON.
func RunRound(ctx context.Context, round Round, screen io.Writer) (Result, error) {
	command := exec.CommandContext(ctx, claudeBinary(),
		"-p", round.Prompt,
		"--model", "opus",
		"--max-turns", strconv.Itoa(round.MaxTurns),
		"--output-format", "stream-json", "--verbose",
		"--allowedTools", strings.Join(round.AllowedTools, ","),
	)
	if round.SessionID != "" {
		command.Args = append(command.Args, "--resume", round.SessionID)
	}
	command.Dir = round.WorktreeDir
	if round.AuthorName != "" && round.AuthorEmail != "" {
		command.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME="+round.AuthorName, "GIT_AUTHOR_EMAIL="+round.AuthorEmail,
			"GIT_COMMITTER_NAME="+round.AuthorName, "GIT_COMMITTER_EMAIL="+round.AuthorEmail,
		)
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return Result{}, err
	}

	if err := os.MkdirAll(filepath.Dir(round.LogPath), 0o755); err != nil {
		return Result{}, err
	}
	logFile, err := os.Create(round.LogPath)
	if err != nil {
		return Result{}, err
	}
	defer logFile.Close()

	if err := command.Start(); err != nil {
		return Result{}, fmt.Errorf("claude: %w", err)
	}
	lastLine, copyErr := copyLinesToLogAndScreen(stdoutPipe, logFile, screen, round.WorktreeDir)
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
		return Result{}, fmt.Errorf("claude: %w (log: %s)", runErr, round.LogPath)
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
