package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// oneTarget makes a temp folder holding a git repo with a bare origin, a one-unit plan and a nightcap.yml
// pointing at both, swaps in a fake gh with no PR, moves into it, and returns the repo path.
func oneTarget(t *testing.T) string {
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
	planText := []byte(strings.Join([]string{
		"## feat: feed (feed)",
		"",
		"- [ ] add it",
		"  - DONE  go test passes",
	}, "\n") + "\n")
	if err := os.WriteFile("plan.md", planText, 0o644); err != nil {
		t.Fatal(err)
	}
	yml := []byte(strings.Join([]string{
		"targets:",
		"  - plan: plan.md",
		"    repo: " + repo,
		"    check: true",
	}, "\n") + "\n")
	if err := os.WriteFile("nightcap.yml", yml, 0o644); err != nil {
		t.Fatal(err)
	}
	useFakeGh(t)
	return repo
}

// useFakeClaude points NIGHTCAP_CLAUDE at a script that answers every round with a Done result JSON.
func useFakeClaude(t *testing.T) {
	t.Helper()
	fakeClaude := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\necho '{\"type\":\"result\",\"result\":\"ok\"}'\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)
}

// useFakeGh points NIGHTCAP_GH at a script that runs viewLines for `gh pr view`, or answers "no PR" when there are none,
// and prints one PR URL for anything else, the way `gh pr create` does.
func useFakeGh(t *testing.T, viewLines ...string) {
	t.Helper()
	if len(viewLines) == 0 {
		viewLines = []string{`echo 'no pull requests found for branch "feat/feed"' >&2`, "exit 1"}
	}
	fakeGh := filepath.Join(t.TempDir(), "gh")
	script := strings.Join(slices.Concat(
		[]string{"#!/bin/sh", `if [ "$2" = "view" ]; then`},
		viewLines,
		[]string{"fi", "echo https://github.com/foreverfl/feed/pull/1"},
	), "\n")
	if err := os.WriteFile(fakeGh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_GH", fakeGh)
}

func TestDispatchRunPrintsPickedSection(t *testing.T) {
	repo := oneTarget(t)
	useFakeClaude(t)
	var stdout, stderr bytes.Buffer
	code := dispatch([]string{"run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	want := strings.Join([]string{
		"plan.md",
		"  branch: feat/feed",
		"  worktree: " + filepath.Join(repo, ".nightcap/worktrees/feed"),
		"  take  add it",
		"  ran  add it: done (log: " + filepath.Join(repo, ".nightcap/logs/feed/unit1-round1.log") + ")",
		"  check: green",
		"  added lines: 0",
		"  verdict: green",
		"  draft PR: https://github.com/foreverfl/feed/pull/1",
		"  ready: feat/feed",
		"  label: nightcap:read-pr",
	}, "\n") + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunMarksUnitBlockedAfterMaxRounds(t *testing.T) {
	repo := oneTarget(t)
	useFakeClaude(t)
	yml := []byte(strings.Join([]string{
		"max_rounds: 2",
		"targets:",
		"  - plan: plan.md",
		"    repo: " + repo,
		"    check: echo $$; false",
	}, "\n"))
	if err := os.WriteFile("nightcap.yml", yml, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	logs, _ := filepath.Glob(filepath.Join(repo, ".nightcap/logs/feed/*.log"))
	if len(logs) != 2 {
		t.Fatalf("max_rounds 2 should leave 2 round logs, got %d: %v", len(logs), logs)
	}
	if !strings.HasSuffix(stdout.String(), "  check: red\n  blocked\n") {
		t.Fatalf("stdout should end with a red check and blocked, got %q", stdout.String())
	}
}

func TestRunMarksUnitBlockedWhenTheSameFailureRepeats(t *testing.T) {
	repo := oneTarget(t)
	useFakeClaude(t)
	yml := []byte(strings.Join([]string{
		"max_rounds: 3",
		"targets:",
		"  - plan: plan.md",
		"    repo: " + repo,
		"    check: false",
	}, "\n"))
	if err := os.WriteFile("nightcap.yml", yml, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	logs, _ := filepath.Glob(filepath.Join(repo, ".nightcap/logs/feed/*.log"))
	if len(logs) != 2 {
		t.Fatalf("`false` fails the same way every round, so it should stop after 2 of the 3 rounds; got %d logs: %v", len(logs), logs)
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

func TestRunMarksUnitBlockedWhenTheSameVerdictRepeats(t *testing.T) {
	repo := oneTarget(t)
	fakeClaude := filepath.Join(t.TempDir(), "claude")
	script := strings.Join([]string{
		"#!/bin/sh",
		"echo $$ >> go.sum",
		"git add go.sum",
		"git -c user.name=t -c user.email=t@t commit -q -m 'touch go.sum'",
		`echo '{"type":"result","result":"ok"}'`,
	}, "\n")
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)
	yml := []byte(strings.Join([]string{
		"max_rounds: 3",
		"targets:",
		"  - plan: plan.md",
		"    repo: " + repo,
		"    check: true",
	}, "\n"))
	if err := os.WriteFile("nightcap.yml", yml, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	logs, _ := filepath.Glob(filepath.Join(repo, ".nightcap/logs/feed/*.log"))
	if len(logs) != 2 {
		t.Fatalf("go.sum turns the verdict red the same way every round, so it should stop after 2 of the 3 rounds; got %d logs: %v", len(logs), logs)
	}
	if !strings.HasSuffix(stdout.String(), "  verdict: red (changed a dependency file: go.sum)\n  blocked\n") {
		t.Fatalf("stdout should end with a red verdict and blocked, got %q", stdout.String())
	}
}

func TestRunSkipsAHeadingThatEndedBlocked(t *testing.T) {
	repo := oneTarget(t)
	useFakeClaude(t)
	yml := []byte(strings.Join([]string{
		"max_rounds: 2",
		"targets:",
		"  - plan: plan.md",
		"    repo: " + repo,
		"    check: false",
	}, "\n"))
	if err := os.WriteFile("nightcap.yml", yml, 0o644); err != nil {
		t.Fatal(err)
	}
	var firstStdout, secondStdout, stderr bytes.Buffer
	if code := dispatch([]string{"run"}, &firstStdout, &stderr); code != 0 {
		t.Fatalf("first run: exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	if code := dispatch([]string{"run"}, &secondStdout, &stderr); code != 0 {
		t.Fatalf("second run: exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	want := strings.Join([]string{
		"skip feed: blocked",
		"plan.md: nothing to take",
	}, "\n") + "\n"
	if secondStdout.String() != want {
		t.Fatalf("second run stdout = %q, want %q", secondStdout.String(), want)
	}
}

func TestRunSkipsAHeadingWhosePRWasMerged(t *testing.T) {
	oneTarget(t)
	useFakeGh(t, `echo '{"state":"MERGED","url":"https://github.com/foreverfl/feed/pull/1"}'`, "exit 0")
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	want := strings.Join([]string{
		"skip feed: PR merged",
		"plan.md: nothing to take",
	}, "\n") + "\n"
	if stdout.String() != want {
		t.Fatalf("a heading whose PR was merged should be skipped, got %q", stdout.String())
	}
}
