package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// oneConfig is the smallest valid nightcap.yml: one target with only the required keys.
func oneConfig() string {
	return "targets:\n  - plan: docs/plan.md\n    repo: /tmp/repo\n    check: make check\n"
}

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
	got, err := loadFrom(t, oneConfig())
	if err != nil {
		t.Fatal(err)
	}
	want := Config{MaxTurns: 30, MaxRounds: 3, AllowedTools: defaultAllowedTools, Lang: []string{"ko"}, MergeMethod: "merge",
		Targets: []Target{{Plan: "docs/plan.md", Repo: "/tmp/repo", Check: "make check",
			Protected: DefaultProtected, Dependencies: DefaultDependencies}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoadOverridesMaxTurns(t *testing.T) {
	got, err := loadFrom(t, "max_turns: 5\n"+oneConfig())
	if err != nil {
		t.Fatal(err)
	}
	if got.MaxTurns != 5 {
		t.Fatalf("MaxTurns = %d, want 5", got.MaxTurns)
	}
}

func TestLoadExpandsHomeInRepo(t *testing.T) {
	got, err := loadFrom(t, "targets:\n  - plan: docs/plan.md\n    repo: ~/repo\n    check: make check\n")
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	if got.Targets[0].Repo != filepath.Join(home, "repo") {
		t.Fatalf("Repo = %q, want %q", got.Targets[0].Repo, filepath.Join(home, "repo"))
	}
}

func TestLoadRequiresTargets(t *testing.T) {
	if _, err := loadFrom(t, "max_turns: 5\n"); err == nil {
		t.Fatal("expected error for missing targets")
	}
}

func TestLoadRequiresPlanInTarget(t *testing.T) {
	if _, err := loadFrom(t, "targets:\n  - repo: /tmp/repo\n    check: make check\n"); err == nil {
		t.Fatal("expected error for target without plan")
	}
}

func TestLoadReadsLang(t *testing.T) {
	got, err := loadFrom(t, "lang: [en, ko]\n"+oneConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Lang, []string{"en", "ko"}) {
		t.Fatalf("Lang = %v, want [en ko]", got.Lang)
	}
}

func TestLoadReadsWorkStart(t *testing.T) {
	got, err := loadFrom(t, "work_start: \"12:00\"\n"+oneConfig())
	if err != nil {
		t.Fatal(err)
	}
	if got.WorkStart != "12:00" {
		t.Fatalf("WorkStart = %q, want 12:00", got.WorkStart)
	}
}

func TestLoadReadsLimits(t *testing.T) {
	got, err := loadFrom(t, "limit_session: 80\nlimit_weekly_all: 60\n"+oneConfig())
	if err != nil {
		t.Fatal(err)
	}
	if got.LimitSession != 80 || got.LimitWeeklyAll != 60 {
		t.Fatalf("LimitSession, LimitWeeklyAll = %d, %d, want 80, 60", got.LimitSession, got.LimitWeeklyAll)
	}
}

func TestLoadReadsProtected(t *testing.T) {
	got, err := loadFrom(t, oneConfig()+"    protected: [\"*.tsp\", \"*.hurl\"]\n")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Targets[0].Protected, []string{"*.tsp", "*.hurl"}) {
		t.Fatalf("Protected = %v, want [*.tsp *.hurl]", got.Targets[0].Protected)
	}
}
