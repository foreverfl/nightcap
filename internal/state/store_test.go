package state

import (
	"path/filepath"
	"testing"
	"time"
)

// oneState returns a State with one PR holding one commit that ended green.
func oneState() State {
	return State{PRs: map[string]PR{
		"feed": {
			Commits: map[string]Commit{
				"add-it": {Attempts: 1, Result: "green", UpdatedAt: time.Date(2026, 11, 3, 2, 0, 0, 0, time.UTC)},
			},
		},
	}}
}

func TestSaveThenLoadGivesTheSameCommit(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".nightcap", "state.json")
	if err := Save(path, oneState()); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := oneState().PRs["feed"].Commits["add-it"]
	if got := loaded.PRs["feed"].Commits["add-it"]; got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestLoadWithoutFileIsEmpty(t *testing.T) {
	loaded, err := Load(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.PRs) != 0 {
		t.Fatalf("a missing file should give no PRs, got %+v", loaded.PRs)
	}
}
