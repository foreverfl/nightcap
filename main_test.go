package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"nightcap/internal/testsetup"
)

func TestDispatchRunPrintsPickedPR(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeClaude(t)
	var stdout, stderr bytes.Buffer
	code := dispatch([]string{"run", "commit-bot.md"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	want := strings.Join([]string{
		"commit-bot.md",
		"  branch: feat/feed",
		"  worktree: " + filepath.Join(testsetup.RunCacheDir(t, repo), "worktrees/feed"),
		"  take  add it",
		"  ran  add it: done (log: " + filepath.Join(testsetup.RunDataDir(t, repo), "logs/feed/commit1-round1.log") + ")",
		"  test: green",
		"  added lines: 0",
		"  verdict: green",
		"  PR: https://github.com/foreverfl/feed/pull/1",
		"  label: nightcap:read-pr",
	}, "\n") + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q", stdout.String())
	}
	wantKinds := []string{"run_start", "commit_start", "round", "test", "verdict", "pr_open", "run_end"}
	if kinds := testsetup.RunFileKinds(t, repo); !slices.Equal(kinds, wantKinds) {
		t.Fatalf("kinds = %v, want %v", kinds, wantKinds)
	}
}

func TestRunMarksCommitBlockedAfterMaxRounds(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeClaude(t)
	testsetup.WritePlanAndConfig(t, repo, "echo $$; false", "max_rounds: 2\n")
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run", "commit-bot.md"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	logs, _ := filepath.Glob(filepath.Join(testsetup.RunDataDir(t, repo), "logs/feed/*.log"))
	if len(logs) != 2 {
		t.Fatalf("max_rounds 2 should leave 2 round logs, got %d: %v", len(logs), logs)
	}
	if !strings.HasSuffix(stdout.String(), "  test: red\n  blocked\n") {
		t.Fatalf("stdout should end with a red check and blocked, got %q", stdout.String())
	}
}

func TestRunLogsEachRoundAndTestToTheRunFile(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeClaude(t)
	testsetup.WritePlanAndConfig(t, repo, "echo $$; false", "max_rounds: 2\n")
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run", "commit-bot.md"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	want := []string{"run_start", "commit_start", "round", "test", "round", "test", "blocked", "run_end"}
	if kinds := testsetup.RunFileKinds(t, repo); !slices.Equal(kinds, want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
}

func TestRunMarksCommitBlockedWhenTheSameFailureRepeats(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeClaude(t)
	testsetup.WritePlanAndConfig(t, repo, "false", "max_rounds: 3\n")
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run", "commit-bot.md"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	logs, _ := filepath.Glob(filepath.Join(testsetup.RunDataDir(t, repo), "logs/feed/*.log"))
	if len(logs) != 2 {
		t.Fatalf("`false` fails the same way every round, so it should stop after 2 of the 3 rounds; got %d logs: %v", len(logs), logs)
	}
}

func TestDispatchVersionPrintsTheBuiltInVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}
	if stdout.String() != "nightcap dev\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "nightcap dev\n")
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

func TestRunMarksCommitBlockedWhenTheSameVerdictRepeats(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
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
	testsetup.WritePlanAndConfig(t, repo, "true", "max_rounds: 3\n")
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run", "commit-bot.md"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	logs, _ := filepath.Glob(filepath.Join(testsetup.RunDataDir(t, repo), "logs/feed/*.log"))
	if len(logs) != 2 {
		t.Fatalf("go.sum turns the verdict red the same way every round, so it should stop after 2 of the 3 rounds; got %d logs: %v", len(logs), logs)
	}
	if !strings.HasSuffix(stdout.String(), "  verdict: red (changed a dependency file: go.sum)\n  blocked\n") {
		t.Fatalf("stdout should end with a red verdict and blocked, got %q", stdout.String())
	}
	want := []string{"run_start", "commit_start", "round", "test", "verdict", "round", "test", "verdict", "blocked", "run_end"}
	if kinds := testsetup.RunFileKinds(t, repo); !slices.Equal(kinds, want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
}

func TestRunStopsAtAPRThatEndedBlocked(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeClaude(t)
	testsetup.WritePlanAndConfig(t, repo, "false", "max_rounds: 2\n")
	planFile, err := os.OpenFile("commit-bot.md", os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	// a second PR below feed, which must wait while feed is blocked
	secondPR := strings.Join([]string{
		"",
		"## 📄 feat: feed page (feed-page)",
		"",
		"- [ ] show it",
		"  - DONE  go test passes",
	}, "\n") + "\n"
	if _, err := planFile.WriteString(secondPR); err != nil {
		t.Fatal(err)
	}
	planFile.Close()
	var firstStdout, secondStdout, stderr bytes.Buffer
	if code := dispatch([]string{"run", "commit-bot.md"}, &firstStdout, &stderr); code != 0 {
		t.Fatalf("first run: exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	if code := dispatch([]string{"run", "commit-bot.md"}, &secondStdout, &stderr); code != 0 {
		t.Fatalf("second run: exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	want := strings.Join([]string{
		"stop feed: blocked, the PRs below wait",
		"commit-bot.md: nothing to take",
	}, "\n") + "\n"
	if secondStdout.String() != want {
		t.Fatalf("second run stdout = %q, want %q", secondStdout.String(), want)
	}
}

func TestRunSkipsAPRThatWasMerged(t *testing.T) {
	testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeGh(t, map[string]string{"feat/feed": "MERGED"})
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run", "commit-bot.md"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	want := strings.Join([]string{
		"skip feed: PR merged",
		"commit-bot.md: nothing to take",
	}, "\n") + "\n"
	if stdout.String() != want {
		t.Fatalf("a PR that was merged should be skipped, got %q", stdout.String())
	}
}

func TestRunWithoutAPlanFileFails(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run"}, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "run: name a plan file") {
		t.Fatalf("want exit 1 and run: name a plan file, got %d %q", code, stderr.String())
	}
}

func TestRunWithoutProtectedInTheFrontmatterFails(t *testing.T) {
	t.Chdir(t.TempDir())
	planText := strings.Join([]string{
		"---",
		"repo: /tmp/repo",
		"test: true",
		"dependencies: []",
		"---",
	}, "\n") + "\n"
	if err := os.WriteFile("commit-bot.md", []byte(planText), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	want := "run: the plan's frontmatter needs repo, test, protected and dependencies"
	if code := dispatch([]string{"run", "commit-bot.md"}, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), want) {
		t.Fatalf("want exit 1 and %q, got %d %q", want, code, stderr.String())
	}
}

func TestRunStartsOnThePreviousPRsOpenBranch(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeClaude(t)
	planText := []byte(strings.Join([]string{
		"---",
		"repo: " + repo,
		"test: true",
		"protected: []",
		"dependencies: []",
		"---",
		"",
		"## 📄 feat: feed (feed)",
		"",
		"- [ ] add it",
		"",
		"## 📄 feat: feed page (feed-page)",
		"",
		"- [ ] show it",
		"  - DONE  go test passes",
	}, "\n") + "\n")
	if err := os.WriteFile("commit-bot.md", planText, 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("git", "-C", repo, "branch", "feat/feed").CombinedOutput(); err != nil {
		t.Fatalf("git branch: %v\n%s", err, output)
	}
	ghCallsPath := testsetup.UseFakeGh(t, map[string]string{"feat/feed": "OPEN"})
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run", "commit-bot.md"}, &stdout, &stderr); code != 0 {
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

func TestRunResumesTheSessionOnTheSecondRound(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	// the fake prints its arguments to stderr, which lands in the round log, and answers with one session_id
	fakeClaude := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\necho \"$@\" >&2\necho '{\"type\":\"result\",\"result\":\"ok\",\"session_id\":\"s1\"}'\n"
	if err := os.WriteFile(fakeClaude, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NIGHTCAP_CLAUDE", fakeClaude)
	testsetup.WritePlanAndConfig(t, repo, "echo $$; false", "max_rounds: 2\n")
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"run", "commit-bot.md"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	first, _ := os.ReadFile(filepath.Join(testsetup.RunDataDir(t, repo), "logs/feed/commit1-round1.log"))
	second, _ := os.ReadFile(filepath.Join(testsetup.RunDataDir(t, repo), "logs/feed/commit1-round2.log"))
	if strings.Contains(string(first), "--resume") || !strings.Contains(string(second), "--resume s1") {
		t.Fatalf("round 1 should start a session and round 2 resume s1\nround 1: %s\nround 2: %s", first, second)
	}
}
