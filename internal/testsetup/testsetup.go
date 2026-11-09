// Package testsetup builds the temp repo, plan, config and fake claude and gh that the nightcap run tests share.
package testsetup

import (
	"encoding/json"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"nightcap/internal/path"
)

// SetUpRepoWithOneCommitPlan moves into a temp folder with a git repo, a one-commit plan and a fake gh, and returns the repo path.
func SetUpRepoWithOneCommitPlan(t *testing.T) string {
	t.Helper()
	t.Chdir(t.TempDir())
	repo, _ := filepath.Abs("repo")
	origin, _ := filepath.Abs("origin.git")
	for _, args := range [][]string{
		{"init", "-q", "-b", "main", repo},
		{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
		{"init", "-q", "--bare", origin},
		{"-C", repo, "remote", "add", "origin", origin},
	} {
		if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	WritePlanAndConfig(t, repo, "true", "")
	UseFakeGh(t, nil)
	return repo
}

// WritePlan writes commit-bot.md: a frontmatter naming repo, the test command and the protected and dependency files, then one PR with one commit.
func WritePlan(t *testing.T, repo, test string) {
	t.Helper()
	planText := strings.Join([]string{
		"---",
		"repo: " + repo,
		"test: " + test,
		`protected: ["*_test.go"]`,
		"dependencies: [go.mod, go.sum]",
		"---",
		"",
		"## 📄 feat: first (first)",
		"",
		"- [ ] add it",
		"  - DONE  go test passes",
	}, "\n") + "\n"
	if err := os.WriteFile("commit-bot.md", []byte(planText), 0o644); err != nil {
		t.Fatal(err)
	}
}

// WritePlanAndConfig writes commit-bot.md through WritePlan, then nightcap.yml with configYAML under a temp XDG_CONFIG_HOME.
func WritePlanAndConfig(t *testing.T, repo, test, configYAML string) {
	t.Helper()
	WritePlan(t, repo, test)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	if err := os.MkdirAll(filepath.Join(configHome, "nightcap"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configHome, "nightcap", "nightcap.yml"), []byte(configYAML), 0o644); err != nil {
		t.Fatal(err)
	}
}

// UseFakeClaude points NIGHTCAP_CLAUDE at a script that answers every round with a Done result JSON.
func UseFakeClaude(t *testing.T) {
	t.Helper()
	fakeClaude := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\necho '{\"type\":\"result\",\"result\":\"ok\"}'\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)
}

// fakeGhPath is the repo's testdata/fake-gh.sh, found from this file so it works from any package's tests.
var fakeGhPath = func() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "fake-gh.sh")
}()

// UseFakeGh swaps in a fake gh that answers `gh pr view <branch>` from prStates and prints one PR URL for anything else.
// It returns the file where the fake writes each call's arguments, one call per line.
func UseFakeGh(t *testing.T, prStates map[string]string) string {
	t.Helper()
	var stateLines []string
	for _, branch := range slices.Sorted(maps.Keys(prStates)) {
		stateLines = append(stateLines, branch+" "+prStates[branch]+"\n")
	}
	statesPath := filepath.Join(t.TempDir(), "gh-states")
	if err := os.WriteFile(statesPath, []byte(strings.Join(stateLines, "")), 0o644); err != nil {
		t.Fatal(err)
	}
	callsPath := filepath.Join(t.TempDir(), "gh-calls")
	t.Setenv("NIGHTCAP_GH", fakeGhPath)
	t.Setenv("FAKE_GH_STATES", statesPath)
	t.Setenv("FAKE_GH_CALLS", callsPath)
	return callsPath
}

// RunFileKinds reads the one run file a run left under repo's data folder and returns the kind of each line in order.
func RunFileKinds(t *testing.T, repo string) []string {
	t.Helper()
	runFiles, _ := filepath.Glob(filepath.Join(RunDataDir(t, repo), "runs/*/*.jsonl"))
	if len(runFiles) != 1 {
		t.Fatalf("one run should leave one run file, got %v", runFiles)
	}
	text, err := os.ReadFile(runFiles[0])
	if err != nil {
		t.Fatal(err)
	}
	var kinds []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(text)), "\n") {
		var event struct{ Kind string }
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, event.Kind)
	}
	return kinds
}

// RunDataDir returns the folder a run writes repo's state, run files and logs to.
func RunDataDir(t *testing.T, repo string) string {
	t.Helper()
	dir, err := path.DataDir(repo)
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// RunCacheDir returns the folder a run puts repo's worktrees under.
func RunCacheDir(t *testing.T, repo string) string {
	t.Helper()
	dir, err := path.CacheDir(repo)
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// AppendToPlan adds lines to the end of commit-bot.md, one per line, after a blank line.
func AppendToPlan(t *testing.T, lines ...string) {
	t.Helper()
	planFile, err := os.OpenFile("commit-bot.md", os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer planFile.Close()
	if _, err := planFile.WriteString("\n" + strings.Join(lines, "\n") + "\n"); err != nil {
		t.Fatal(err)
	}
}
