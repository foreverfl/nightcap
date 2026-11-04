package claude

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewOrderSendsTheDiffAndReturnsTheAnswer(t *testing.T) {
	folder := t.TempDir()
	promptPath := filepath.Join(folder, "prompt.txt")
	fakeClaude := filepath.Join(folder, "claude")
	script := strings.Join([]string{
		"#!/bin/sh",
		`printf '%s' "$2" > "` + promptPath + `"`,
		`echo '{"type":"result","result":"흐름 1. a: b"}'`,
	}, "\n")
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)
	answer, err := ReviewOrder(context.Background(), "+func LinkStack(")
	if err != nil {
		t.Fatal(err)
	}
	if answer != "흐름 1. a: b" {
		t.Fatalf("the answer should be the result text %q, got %q", "흐름 1. a: b", answer)
	}
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(prompt), "+func LinkStack(") {
		t.Fatalf("the prompt should hold the diff, got %q", prompt)
	}
}
