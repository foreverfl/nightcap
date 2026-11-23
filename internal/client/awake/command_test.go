package awake

import (
	"slices"
	"testing"
)

func TestCommandOnMacIsCaffeinateWaitingForThePid(t *testing.T) {
	want := []string{"caffeinate", "-i", "-w", "4821"}
	if got := Command("darwin", 4821); !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCommandOnLinuxIsSystemdInhibitHoldingATailOnThePid(t *testing.T) {
	want := []string{"systemd-inhibit", "--what=sleep:idle", "--who=nightcap", "--why=nightcap run",
		"tail", "--pid=4821", "-f", "/dev/null"}
	if got := Command("linux", 4821); !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCommandOnWindowsIsNil(t *testing.T) {
	if got := Command("windows", 4821); got != nil {
		t.Fatalf("got %q, want nil", got)
	}
}
