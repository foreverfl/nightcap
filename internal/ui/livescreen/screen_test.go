package livescreen

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestScreenDrawsTheLinesItGotThenTheElapsedTime(t *testing.T) {
	started := time.Date(2026, 11, 11, 23, 16, 35, 0, time.UTC)
	var current tea.Model = screen{started: started, now: started.Add(3*time.Minute + 12*time.Second)}
	current, _ = current.Update(printedLineMsg("take  add it"))
	current, _ = current.Update(printedLineMsg("ran  add it: done"))
	want := strings.Join([]string{
		"take  add it",
		"ran  add it: done",
		"00:03:12",
	}, "\n")
	if got := current.View().Content; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestScreenStopsTheWorkOnCtrlC(t *testing.T) {
	stopCalls := 0
	var current tea.Model = screen{stopWork: func() { stopCalls++ }}
	current.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if stopCalls != 1 {
		t.Fatalf("Ctrl+C should cancel the work once, got %d", stopCalls)
	}
}
