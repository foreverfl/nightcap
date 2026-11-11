package claude

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nightcap/internal/testsetup"
)

func TestRunRoundWritesEveryStreamLineToTheLogInOrder(t *testing.T) {
	testsetup.UseFakeClaude(t)
	logPath := filepath.Join(t.TempDir(), "commit1-round1.log")

	result, err := RunRound(context.Background(), t.TempDir(), "add it", "", 1, nil, logPath)
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
