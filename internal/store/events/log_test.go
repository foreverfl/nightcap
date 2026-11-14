package events

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendAddsOneLinePerEvent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".nightcap", "runs", "261105", "225500.jsonl")
	at := time.Date(2026, 11, 5, 22, 55, 0, 0, time.UTC)
	if err := Append(path, Event{Kind: "run_start", At: at}); err != nil {
		t.Fatal(err)
	}
	if err := Append(path, Event{Kind: "run_end", At: at, Reason: "session 85% ≥ 80"}); err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		`{"kind":"run_start","at":"2026-11-05T22:55:00Z"}`,
		`{"kind":"run_end","at":"2026-11-05T22:55:00Z","reason":"session 85% ≥ 80"}`,
	}, "\n") + "\n"
	if string(text) != want {
		t.Fatalf("got\n%s\nwant\n%s", text, want)
	}
}
