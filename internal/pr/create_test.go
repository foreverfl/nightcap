package pr

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCreateArgsOpenAPRWithTitleAndBodyFile(t *testing.T) {
	args := createArgs("main", "feat: feed", "/tmp/body.md")
	want := []string{"pr", "create", "--base", "main", "--title", "feat: feed", "--body-file", "/tmp/body.md"}
	if !slices.Equal(args, want) {
		t.Fatalf("want %q, got %q", want, args)
	}
}

func TestCreatePassesTheAppTokenToGh(t *testing.T) {
	dir := t.TempDir()
	fakeGh := filepath.Join(dir, "gh")
	script := strings.Join([]string{
		`#!/bin/sh`,
		`echo "$GH_TOKEN" > "$(dirname "$0")/token"`,
		`echo https://github.com/foreverfl/first/pull/1`,
	}, "\n") + "\n"
	if err := os.WriteFile(fakeGh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_GH", fakeGh)

	if _, err := Create(t.TempDir(), "ghs_fake", "main", "feat: feed", "body"); err != nil {
		t.Fatal(err)
	}

	token, err := os.ReadFile(filepath.Join(dir, "token"))
	if err != nil {
		t.Fatal(err)
	}
	if string(token) != "ghs_fake\n" {
		t.Fatalf("GH_TOKEN = %q, want ghs_fake", token)
	}
}
