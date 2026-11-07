package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// loadFrom writes text as nightcap.yml into a fresh temp dir and loads it.
func loadFrom(t *testing.T, text string) (Config, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return Load(dir)
}

func TestLoadFillsDefaults(t *testing.T) {
	got, err := loadFrom(t, "")
	if err != nil {
		t.Fatal(err)
	}
	want := Config{MaxTurns: 30, MaxRounds: 3, AllowedTools: defaultAllowedTools, Lang: []string{"ko"}, MergeMethod: "merge"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoadOverridesMaxTurns(t *testing.T) {
	got, err := loadFrom(t, "max_turns: 5\n")
	if err != nil {
		t.Fatal(err)
	}
	if got.MaxTurns != 5 {
		t.Fatalf("MaxTurns = %d, want 5", got.MaxTurns)
	}
}

func TestLoadReadsLang(t *testing.T) {
	got, err := loadFrom(t, "lang: [en, ko]\n")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Lang, []string{"en", "ko"}) {
		t.Fatalf("Lang = %v, want [en ko]", got.Lang)
	}
}

func TestLoadReadsWorkStart(t *testing.T) {
	got, err := loadFrom(t, "work_start: \"12:00\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if got.WorkStart != "12:00" {
		t.Fatalf("WorkStart = %q, want 12:00", got.WorkStart)
	}
}

func TestLoadReadsLimits(t *testing.T) {
	got, err := loadFrom(t, "limit_session: 80\nlimit_weekly_all: 60\n")
	if err != nil {
		t.Fatal(err)
	}
	if got.LimitSession != 80 || got.LimitWeeklyAll != 60 {
		t.Fatalf("LimitSession, LimitWeeklyAll = %d, %d, want 80, 60", got.LimitSession, got.LimitWeeklyAll)
	}
}

func TestLoadReadsTheExample(t *testing.T) {
	got, err := Load("../../examples")
	if err != nil {
		t.Fatal(err)
	}
	if got.MergeMethod != "rebase" || !reflect.DeepEqual(got.Lang, []string{"en", "ko"}) {
		t.Fatalf("got MergeMethod %q, Lang %v; want rebase, [en ko]", got.MergeMethod, got.Lang)
	}
}
