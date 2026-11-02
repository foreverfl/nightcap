package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// oneTarget makes a temp folder holding a git repo, a one-unit plan and a nightcap.yml
// pointing at both, moves into it, and returns the repo path.
func oneTarget(t *testing.T) string {
	t.Helper()
	t.Chdir(t.TempDir())
	repo, _ := filepath.Abs("repo")
	for _, args := range [][]string{
		{"init", "-q", "-b", "main", repo},
		{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	planText := []byte("## feat: feed (feed)\n\n- [ ] add it\n  - DONE  go test passes\n")
	if err := os.WriteFile("plan.md", planText, 0o644); err != nil {
		t.Fatal(err)
	}
	yml := []byte("targets:\n  - plan: plan.md\n    repo: " + repo + "\n    check: true\n")
	if err := os.WriteFile("nightcap.yml", yml, 0o644); err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestDispatchRunPrintsPickedSection(t *testing.T) {
	repo := oneTarget(t)
	fakeClaude := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\necho '{\"type\":\"result\",\"result\":\"ok\"}'\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)
	var stdout, stderr bytes.Buffer
	code := dispatch([]string{"run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	want := "plan.md\n  branch: feat/feed\n  worktree: " + filepath.Join(repo, ".nightcap/worktrees/feed") + "\n  take  add it\n  ran  add it: done (log: " + filepath.Join(repo, ".nightcap/logs/feed/unit1-round1.log") + ")\n  check: green\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestDispatchUnknownCommandPrintsUsageAndExits2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := dispatch([]string{"foo"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr.String() != usage {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
