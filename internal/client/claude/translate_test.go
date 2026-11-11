package claude

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTranslateReturnsTheResultText(t *testing.T) {
	fakeClaude := filepath.Join(t.TempDir(), "claude")
	script := strings.Join([]string{
		"#!/bin/sh",
		`echo '{"type":"result","result":"## What"}'`,
	}, "\n")
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)
	english, err := Translate(context.Background(), "## 무엇을")
	if err != nil {
		t.Fatal(err)
	}
	if english != "## What" {
		t.Fatalf("the translation should be the result text %q, got %q", "## What", english)
	}
}

func TestTranslateTitleSendsTheInstructionsAndTheTitle(t *testing.T) {
	folder := t.TempDir()
	promptPath := filepath.Join(folder, "prompt.txt")
	fakeClaude := filepath.Join(folder, "claude")
	script := strings.Join([]string{
		"#!/bin/sh",
		`printf '%s' "$2" > "` + promptPath + `"`,
		`echo '{"type":"result","result":"show the real ranking"}'`,
	}, "\n")
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)

	if _, err := TranslateTitle(context.Background(), "첫 화면이 그날의 실제 순위를 보여준다"); err != nil {
		t.Fatal(err)
	}

	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	want := translateTitleInstructions + "\n\n" + "첫 화면이 그날의 실제 순위를 보여준다"
	if string(prompt) != want {
		t.Fatalf("prompt = %q, want %q", prompt, want)
	}
}
