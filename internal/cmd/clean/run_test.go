package clean

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"nightcap/internal/git"
	"nightcap/internal/testsetup"
)

func TestRunRemovesNightcapWorktree(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	worktreeDir := filepath.Join(testsetup.RunCacheDir(t, repo), "worktrees", "first")
	if _, err := git.AddWorktree(repo, "main", "feat/first", worktreeDir); err != nil {
		t.Fatal(err)
	}

	if err := Run(io.Discard); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(worktreeDir); !os.IsNotExist(err) {
		t.Fatalf("worktree folder still there: %v", err)
	}
	worktrees, err := git.ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(worktrees) != 1 {
		t.Fatalf("worktree list = %v", worktrees)
	}
}

func TestRunWithNoWorktreeSaysNothingToClean(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	var output strings.Builder

	if err := Run(&output); err != nil {
		t.Fatal(err)
	}

	if output.String() != "nothing to clean\n" {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRunDeletesTheWorktreeBranch(t *testing.T) {
	repo := testsetup.SetUpRepoWithOneCommitPlan(t)
	worktreeDir := filepath.Join(testsetup.RunCacheDir(t, repo), "worktrees", "first")
	if _, err := git.AddWorktree(repo, "main", "feat/first", worktreeDir); err != nil {
		t.Fatal(err)
	}

	if err := Run(io.Discard); err != nil {
		t.Fatal(err)
	}

	branches, err := exec.Command("git", "-C", repo, "branch", "--list", "feat/first").Output()
	if err != nil {
		t.Fatal(err)
	}
	if string(branches) != "" {
		t.Fatalf("git branch still lists %q", branches)
	}
}

func TestRunRemovesStateFileAndKeepsLogs(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	repoDataDir := filepath.Join(dataHome, "nightcap", "blog-v2")
	if err := os.MkdirAll(filepath.Join(repoDataDir, "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"state.json", "logs/x.log"} {
		if err := os.WriteFile(filepath.Join(repoDataDir, file), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := Run(io.Discard); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(repoDataDir, "state.json")); !os.IsNotExist(err) {
		t.Fatalf("state.json still there: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repoDataDir, "logs", "x.log")); err != nil {
		t.Fatalf("logs/x.log should stay: %v", err)
	}
}
