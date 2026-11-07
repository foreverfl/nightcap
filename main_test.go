package main

import (
	"bytes"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// setUpRepoWithOneUnitPlan makes a temp folder holding a git repo with a bare origin, a one-unit plan and a nightcap.yml
// pointing at both, swaps in a fake gh with no PR, moves into it, and returns the repo path.
func setUpRepoWithOneUnitPlan(t *testing.T) string {
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
		"## 📄 feat: feed (feed)",
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
	useFakeGh(t, nil)
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

// fakeGhPath is resolved before any test moves into its temp folder.
var fakeGhPath, _ = filepath.Abs("testdata/fake-gh.sh")

// useFakeGh points NIGHTCAP_GH at testdata/fake-gh.sh, which answers `gh pr view <branch>` with that branch's state
// from prStates (OPEN, MERGED or CLOSED), or "no PR" for a branch not in it, and prints one PR URL for anything else.
// It returns the path of the file where the fake writes every call's arguments, one call per line.
func useFakeGh(t *testing.T, prStates map[string]string) string {
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

func TestDispatchRunPrintsPickedSection(t *testing.T) {
	repo := setUpRepoWithOneUnitPlan(t)
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
		"  PR: https://github.com/foreverfl/feed/pull/1",
		"  label: nightcap:read-pr",
	}, "\n") + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunMarksUnitBlockedAfterMaxRounds(t *testing.T) {
	repo := setUpRepoWithOneUnitPlan(t)
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
	repo := setUpRepoWithOneUnitPlan(t)
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
	repo := setUpRepoWithOneUnitPlan(t)
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
	repo := setUpRepoWithOneUnitPlan(t)
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
	setUpRepoWithOneUnitPlan(t)
	useFakeGh(t, map[string]string{"feat/feed": "MERGED"})
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

func TestRunStartsOnThePreviousHeadingsOpenPRBranch(t *testing.T) {
	repo := setUpRepoWithOneUnitPlan(t)
	useFakeClaude(t)
	planText := []byte(strings.Join([]string{
		"## 📄 feat: feed (feed)",
		"",
		"- [ ] add it",
		"  - DONE  go test passes",
		"",
		"## 📄 feat: feed page (feed-page)",
		"",
		"- [ ] show it",
		"  - DONE  go test passes",
	}, "\n") + "\n")
	if err := os.WriteFile("plan.md", planText, 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("git", "-C", repo, "branch", "feat/feed").CombinedOutput(); err != nil {
		t.Fatalf("git branch: %v\n%s", err, output)
	}
	ghCallsPath := useFakeGh(t, map[string]string{"feat/feed": "OPEN"})
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run", "--section", "feed-page"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	ghCalls, err := os.ReadFile(ghCallsPath)
	if err != nil {
		t.Fatal(err)
	}
	createIndex := strings.Index(string(ghCalls), "pr create ")
	if createIndex < 0 || !strings.HasPrefix(string(ghCalls[createIndex:]), "pr create --base feat/feed ") {
		t.Fatalf("the PR should be opened onto feat/feed, got gh calls:\n%s", ghCalls)
	}
}

// oneLintPlan is the smallest plan lint passes: one typed heading with one unit that has a check.
func oneLintPlan() string {
	return "## 📄 feat: a (b)\n" +
		"- [ ] read the docs\n" +
		"  - DONE  go test ./... passes\n"
}

// runLint writes plan to a temp file, runs nightcap lint on it, and returns the exit code and stdout.
func runLint(t *testing.T, plan string) (int, string) {
	t.Helper()
	planPath := filepath.Join(t.TempDir(), "plan.md")
	if err := os.WriteFile(planPath, []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := dispatch([]string{"lint", planPath}, &stdout, &stderr)
	return code, stdout.String()
}

func TestLintPlanWithoutTypedHeadingExits1(t *testing.T) {
	code, stdout := runLint(t, strings.Replace(oneLintPlan(), "## 📄 feat: a (b)", "## cmd", 1))
	if code != 1 || stdout != "no heading with type: slug\n" {
		t.Fatalf("want exit 1 and the no-heading line, got %d %q", code, stdout)
	}
}

func TestLintUnitWithoutCheckExits1(t *testing.T) {
	code, stdout := runLint(t, strings.Replace(oneLintPlan(), "go test ./... passes", "테스트가 없음", 1))
	want := "heading: feat: a (b)\n  no check: read the docs\n"
	if code != 1 || stdout != want {
		t.Fatalf("want exit 1 and %q, got %d %q", want, code, stdout)
	}
}

func TestLintPrintsAPersonHeadingAsPerson(t *testing.T) {
	code, stdout := runLint(t, strings.Replace(oneLintPlan(), "## 📄 feat:", "## 👤 feat:", 1))
	if code != 0 || stdout != "person: feat: a (b)\n" {
		t.Fatalf("want exit 0 and %q, got %d %q", "person: feat: a (b)\n", code, stdout)
	}
}

func TestLintHeadingWithoutAMarkExits1(t *testing.T) {
	code, stdout := runLint(t, strings.Replace(oneLintPlan(), "📄 ", "", 1))
	if code != 1 || stdout != "no mark: feat: a (b)\n" {
		t.Fatalf("want exit 1 and %q, got %d %q", "no mark: feat: a (b)\n", code, stdout)
	}
}

func TestLintWherePathWithASpaceExits1(t *testing.T) {
	code, stdout := runLint(t, strings.Replace(oneLintPlan(), "  - DONE", "  - WHERE (blog-v2) a b/\n  - DONE", 1))
	want := "heading: feat: a (b)\n  WHERE path not readable: a b/\n"
	if code != 1 || stdout != want {
		t.Fatalf("want exit 1 and %q, got %d %q", want, code, stdout)
	}
}

func TestRunResumesTheSessionOnTheSecondRound(t *testing.T) {
	repo := setUpRepoWithOneUnitPlan(t)
	// the fake prints its arguments to stderr, which lands in the round log, and answers with one session_id
	fakeClaude := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\necho \"$@\" >&2\necho '{\"type\":\"result\",\"result\":\"ok\",\"session_id\":\"s1\"}'\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)
	yml := []byte("max_rounds: 2\ntargets:\n  - plan: plan.md\n    repo: " + repo + "\n    check: echo $$; false\n")
	if err := os.WriteFile("nightcap.yml", yml, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	first, _ := os.ReadFile(filepath.Join(repo, ".nightcap/logs/feed/unit1-round1.log"))
	second, _ := os.ReadFile(filepath.Join(repo, ".nightcap/logs/feed/unit1-round2.log"))
	if strings.Contains(string(first), "--resume") || !strings.Contains(string(second), "--resume s1") {
		t.Fatalf("round 1 should start a session and round 2 resume s1\nround 1: %s\nround 2: %s", first, second)
	}
}
