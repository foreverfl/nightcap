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
