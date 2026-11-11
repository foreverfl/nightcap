package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// oneRepo makes an empty git repo with one commit, the least a worktree needs.
func oneRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		if output, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	return repo
}

func TestAddWorktreeCreatesBranchFolder(t *testing.T) {
	repo := oneRepo(t)
	dir, err := AddWorktree(repo, "main", "feat/feed", filepath.Join(repo, ".nightcap", "worktrees", "feed"))
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Join(repo, ".nightcap", "worktrees", "feed") {
		t.Fatalf("dir = %q", dir)
	}
	if !branchExists(repo, "feat/feed") {
		t.Fatal("branch feat/feed was not created")
	}
}

func TestAddWorktreeReusesExistingFolder(t *testing.T) {
	repo := oneRepo(t)
	if _, err := AddWorktree(repo, "main", "feat/feed", filepath.Join(repo, ".nightcap", "worktrees", "feed")); err != nil {
		t.Fatal(err)
	}
	if _, err := AddWorktree(repo, "main", "feat/feed", filepath.Join(repo, ".nightcap", "worktrees", "feed")); err != nil {
		t.Fatalf("second add: %v", err)
	}
}

func TestListWorktreesShowsAddedOne(t *testing.T) {
	repo := oneRepo(t)
	if _, err := AddWorktree(repo, "main", "feat/feed", filepath.Join(repo, ".nightcap", "worktrees", "feed")); err != nil {
		t.Fatal(err)
	}
	lines, err := ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("got %d worktrees, want 2: %q", len(lines), lines)
	}
}

func TestRemoveWorktreeDeletesTheFolder(t *testing.T) {
	repo := oneRepo(t)
	dir, err := AddWorktree(repo, "main", "feat/feed", filepath.Join(t.TempDir(), "feed"))
	if err != nil {
		t.Fatal(err)
	}
	if err := RemoveWorktree(repo, dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("%s should be gone, stat err = %v", dir, err)
	}
	if lines, _ := ListWorktrees(repo); len(lines) != 1 {
		t.Fatalf("got %d worktrees, want 1: %q", len(lines), lines)
	}
}

func TestWorktreeBranchNamesTheCheckedOutBranch(t *testing.T) {
	repo := oneRepo(t)
	dir, err := AddWorktree(repo, "main", "feat/feed", filepath.Join(t.TempDir(), "feed"))
	if err != nil {
		t.Fatal(err)
	}
	branch, err := WorktreeBranch(dir)
	if err != nil {
		t.Fatal(err)
	}
	if branch != "feat/feed" {
		t.Fatalf("branch = %q, want feat/feed", branch)
	}
}

func TestDeleteBranchDeletesAnUnmergedBranch(t *testing.T) {
	repo := oneRepo(t)
	dir, err := AddWorktree(repo, "main", "feat/feed", filepath.Join(t.TempDir(), "feed"))
	if err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("git", "-C", dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "unmerged").CombinedOutput(); err != nil {
		t.Fatalf("commit: %v\n%s", err, output)
	}
	if err := RemoveWorktree(repo, dir); err != nil {
		t.Fatal(err)
	}
	if err := DeleteBranch(repo, "feat/feed"); err != nil {
		t.Fatal(err)
	}
	if branchExists(repo, "feat/feed") {
		t.Fatal("feat/feed should be gone")
	}
}
