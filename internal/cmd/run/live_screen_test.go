package run

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestLiveScreenDrawsTheLinesItGotThenTheElapsedTime(t *testing.T) {
	started := time.Date(2026, 11, 11, 23, 16, 35, 0, time.UTC)
	var screen tea.Model = liveScreen{started: started, now: started.Add(3*time.Minute + 12*time.Second)}
	screen, _ = screen.Update(printedLineMsg("take  add it"))
	screen, _ = screen.Update(printedLineMsg("ran  add it: done"))
	want := strings.Join([]string{
		"take  add it",
		"ran  add it: done",
		"00:03:12",
	}, "\n")
	if got := screen.View().Content; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
