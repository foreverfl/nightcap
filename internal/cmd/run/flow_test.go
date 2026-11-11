package run

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

func TestRunPrintsPickedPR(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeClaude(t)
	var stdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &stdout); err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		"commit-bot.md",
		"  branch: feat/first",
		"  worktree: " + filepath.Join(testsetup.RunCacheDir(t, repo), "worktrees/first"),
		"  take  add it",
		"  ran  add it: done (log: " + filepath.Join(testsetup.RunDataDir(t, repo), "logs/first/commit1-round1.log") + ")",
		"  test: green",
		"  added lines: 0",
		"  verdict: green",
		"  PR: https://github.com/foreverfl/first/pull/1",
		"  label: nightcap:read-pr",
		"commit-bot.md: nothing to take",
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
	var stdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &stdout); err != nil {
		t.Fatal(err)
	}
	logs, _ := filepath.Glob(filepath.Join(testsetup.RunDataDir(t, repo), "logs/first/*.log"))
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
	var stdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &stdout); err != nil {
		t.Fatal(err)
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
	var stdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &stdout); err != nil {
		t.Fatal(err)
	}
	logs, _ := filepath.Glob(filepath.Join(testsetup.RunDataDir(t, repo), "logs/first/*.log"))
	if len(logs) != 2 {
		t.Fatalf("`false` fails the same way every round, so it should stop after 2 of the 3 rounds; got %d logs: %v", len(logs), logs)
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
	var stdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &stdout); err != nil {
		t.Fatal(err)
	}
	logs, _ := filepath.Glob(filepath.Join(testsetup.RunDataDir(t, repo), "logs/first/*.log"))
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

func TestRunTakesAPRThatEndedBlockedAgain(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeClaude(t)
	testsetup.WritePlanAndConfig(t, repo, "false", "max_rounds: 2\n")
	var firstStdout, secondStdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &firstStdout); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if err := Run([]string{"commit-bot.md"}, &secondStdout); err != nil {
		t.Fatalf("second run: %v", err)
	}
	secondRunLines := strings.Split(secondStdout.String(), "\n")
	if secondRunLines[1] != "  branch: feat/first" {
		t.Fatalf("second run should take first again, got %q", secondStdout.String())
	}
}

func TestRunSkipsAPRThatWasMerged(t *testing.T) {
	testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeGh(t, map[string]string{"feat/first": "MERGED"})
	var stdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &stdout); err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		"skip first: PR merged",
		"commit-bot.md: nothing to take",
	}, "\n") + "\n"
	if stdout.String() != want {
		t.Fatalf("a PR that was merged should be skipped, got %q", stdout.String())
	}
}

func TestRunWithoutAPlanFileFails(t *testing.T) {
	var stdout bytes.Buffer
	if err := Run(nil, &stdout); err == nil || !strings.Contains(err.Error(), "run: name a plan file") {
		t.Fatalf("want run: name a plan file, got %v", err)
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
	var stdout bytes.Buffer
	want := "run: the plan's frontmatter needs repo, test, protected and dependencies"
	if err := Run([]string{"commit-bot.md"}, &stdout); err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("want %q, got %v", want, err)
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
		"## 📄 feat: first (first)",
		"",
		"- [ ] add it",
		"",
		"## 📄 feat: second (second)",
		"",
		"- [ ] show it",
		"  - DONE  go test passes",
	}, "\n") + "\n")
	if err := os.WriteFile("commit-bot.md", planText, 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("git", "-C", repo, "branch", "feat/first").CombinedOutput(); err != nil {
		t.Fatalf("git branch: %v\n%s", err, output)
	}
	ghCallsPath := testsetup.UseFakeGh(t, map[string]string{"feat/first": "OPEN"})
	var stdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &stdout); err != nil {
		t.Fatal(err)
	}
	ghCalls, err := os.ReadFile(ghCallsPath)
	if err != nil {
		t.Fatal(err)
	}
	createIndex := strings.Index(string(ghCalls), "pr create ")
	if createIndex < 0 || !strings.HasPrefix(string(ghCalls[createIndex:]), "pr create --base feat/first ") {
		t.Fatalf("the PR should be opened onto feat/first, got gh calls:\n%s", ghCalls)
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
	var stdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &stdout); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(filepath.Join(testsetup.RunDataDir(t, repo), "logs/first/commit1-round1.log"))
	second, _ := os.ReadFile(filepath.Join(testsetup.RunDataDir(t, repo), "logs/first/commit1-round2.log"))
	if strings.Contains(string(first), "--resume") || !strings.Contains(string(second), "--resume s1") {
		t.Fatalf("round 1 should start a session and round 2 resume s1\nround 1: %s\nround 2: %s", first, second)
	}
}

func TestRunRunsEveryOpenCommitOfThePR(t *testing.T) {
	testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeClaude(t)
	ghCalls := testsetup.UseFakeGh(t, nil)
	planFile, err := os.OpenFile("commit-bot.md", os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	// a second commit under first, which the same run must take after the first
	secondCommit := strings.Join([]string{
		"",
		"- [ ] show it",
		"  - DONE  go test passes",
	}, "\n") + "\n"
	if _, err := planFile.WriteString(secondCommit); err != nil {
		t.Fatal(err)
	}
	planFile.Close()
	var stdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &stdout); err != nil {
		t.Fatal(err)
	}
	if ran := strings.Count(stdout.String(), "  ran  "); ran != 2 {
		t.Fatalf("one run should take both commits, got %d in %q", ran, stdout.String())
	}
	calls, err := os.ReadFile(ghCalls)
	if err != nil {
		t.Fatal(err)
	}
	creates, edits := strings.Count(string(calls), "pr create "), strings.Count(string(calls), "pr edit feat/first --body-file")
	if creates != 1 || edits != 1 {
		t.Fatalf("the second commit should edit the PR, not open another: %d create, %d edit\n%s", creates, edits, calls)
	}
}

func TestRunDoesNotRunACommitThatIsAlreadyGreen(t *testing.T) {
	testsetup.SetUpRepoWithOneCommitPlan(t)
	testsetup.UseFakeClaude(t)
	var firstStdout, secondStdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &firstStdout); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if err := Run([]string{"commit-bot.md"}, &secondStdout); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if strings.Contains(secondStdout.String(), "  ran  ") {
		t.Fatalf("the second run should skip the commit the first one finished, got %q", secondStdout.String())
	}
}

func TestRunTakesThePRBelowOnceThePRAboveIsDone(t *testing.T) {
	testsetup.SetUpRepoWithOneCommitPlan(t) // PR first with one commit
	testsetup.UseFakeClaude(t)
	ghCalls := testsetup.UseFakeGh(t, nil)
	testsetup.AppendToPlan(t, "## 📄 feat: second (second)", "", "- [ ] show it", "  - DONE  go test passes")

	var stdout bytes.Buffer
	if err := Run([]string{"commit-bot.md"}, &stdout); err != nil { // one run: first, then second
		t.Fatal(err)
	}

	calls, _ := os.ReadFile(ghCalls)
	for _, want := range []string{"pr create --base main --title feat: first ", "pr create --base feat/first --title feat: second "} {
		if !strings.Contains(string(calls), want) {
			t.Fatalf("gh calls should contain %q\n%s", want, calls)
		}
	}
	if !strings.HasSuffix(stdout.String(), "commit-bot.md: nothing to take\n") {
		t.Fatalf("the run should go on until nothing is left, got %q", stdout.String())
	}
}
