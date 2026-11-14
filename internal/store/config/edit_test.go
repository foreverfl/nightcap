package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditOpensTheConfigInTheEditorWithItsArguments(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	configPath, _, err := Init([]byte("max_rounds: 3\n"))
	if err != nil {
		t.Fatal(err)
	}
	// the fake editor writes the arguments it got into opened
	opened := filepath.Join(t.TempDir(), "opened")
	fakeEditor := filepath.Join(t.TempDir(), "editor")
	if err := os.WriteFile(fakeEditor, []byte("#!/bin/sh\necho \"$@\" > "+opened+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", fakeEditor+" --wait")
	if err := Edit(); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(opened); strings.TrimSpace(string(got)) != "--wait "+configPath {
		t.Fatalf("editor got %q, want --wait %s", got, configPath)
	}
}

func TestEditWithoutAConfigSaysToRunInit(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := Edit(); err == nil || !strings.HasSuffix(err.Error(), "not found, run nightcap init") {
		t.Fatalf("want the run nightcap init error, got %v", err)
	}
}
