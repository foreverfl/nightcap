package claude

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nightcap/internal/testsetup"
)

func TestRunRoundWritesEveryStreamLineToTheLogInOrder(t *testing.T) {
	testsetup.UseFakeClaude(t)
	logPath := filepath.Join(t.TempDir(), "commit1-round1.log")

	result, err := RunRound(context.Background(), Round{WorktreeDir: t.TempDir(), Prompt: "add it", LogPath: logPath}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	if result.Outcome != Done {
		t.Fatalf("outcome = %q, want done", result.Outcome)
	}
	logText, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		`{"type":"system","subtype":"init"}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"ok"}]}}`,
		`{"type":"result","result":"ok"}`,
		"",
		"--- stderr ---",
		"",
	}, "\n")
	if string(logText) != want {
		t.Fatalf("log = %q, want %q", logText, want)
	}
}

func TestRunRoundAsksForOpus(t *testing.T) {
	argsPath := testsetup.UseFakeClaudeRecordingArgs(t)

	_, err := RunRound(context.Background(), Round{WorktreeDir: t.TempDir(), Prompt: "add it", LogPath: filepath.Join(t.TempDir(), "round.log")}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(args), "\n--model\nopus\n") {
		t.Fatalf("args = %q, want --model opus", args)
	}
}

func TestAskOnceAsksForOpus(t *testing.T) {
	argsPath := testsetup.UseFakeClaudeRecordingArgs(t)

	if _, err := askOnce(context.Background(), "test", "say ok"); err != nil {
		t.Fatal(err)
	}

	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(args), "\n--model\nopus\n") {
		t.Fatalf("args = %q, want --model opus", args)
	}
}

func TestRunRoundPassesTheAuthorToClaude(t *testing.T) {
	dir := t.TempDir()
	fakeClaude := filepath.Join(dir, "claude")
	script := strings.Join([]string{
		`#!/bin/sh`,
		`echo "$GIT_AUTHOR_NAME" > "$(dirname "$0")/author"`,
		`echo '{"type":"result","result":"ok"}'`,
	}, "\n") + "\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)

	_, err := RunRound(context.Background(), Round{
		WorktreeDir: t.TempDir(),
		Prompt:      "add it",
		AuthorName:  "nightcap-cli[bot]",
		AuthorEmail: "bot@users.noreply.github.com",
		LogPath:     filepath.Join(dir, "round.log"),
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	author, err := os.ReadFile(filepath.Join(dir, "author"))
	if err != nil {
		t.Fatal(err)
	}
	if string(author) != "nightcap-cli[bot]\n" {
		t.Fatalf("GIT_AUTHOR_NAME = %q, want nightcap-cli[bot]", author)
	}
}
