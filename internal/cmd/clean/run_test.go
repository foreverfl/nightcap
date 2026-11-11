package clean

import (
	"io"
	"os"
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
	var output strings.Builder

	if err := Run(&output); err != nil {
		t.Fatal(err)
	}

	if output.String() != "nothing to clean\n" {
		t.Fatalf("output = %q", output.String())
	}
}
