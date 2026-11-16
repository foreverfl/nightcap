package git

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// oneRepoWithOrigin makes oneRepo with a bare repo as its origin, and returns the repo.
func oneRepoWithOrigin(t *testing.T) string {
	t.Helper()
	repo := oneRepo(t)
	origin := filepath.Join(t.TempDir(), "origin.git")
	for _, args := range [][]string{
		{"init", "-q", "--bare", origin},
		{"-C", repo, "remote", "add", "origin", origin},
	} {
		if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	return repo
}

func TestDeleteRemoteBranchRemovesItFromOrigin(t *testing.T) {
	repo := oneRepoWithOrigin(t)
	if output, err := exec.Command("git", "-C", repo, "push", "-q", "origin", "main:feat/first").CombinedOutput(); err != nil {
		t.Fatalf("git push: %v\n%s", err, output)
	}
	if err := DeleteRemoteBranch(repo, "feat/first"); err != nil {
		t.Fatal(err)
	}
	remoteHeads, err := runGit(repo, "ls-remote", "origin", "feat/first")
	if err != nil {
		t.Fatal(err)
	}
	if remoteHeads != "" {
		t.Fatalf("feat/first should be gone from origin, got %q", remoteHeads)
	}
}

func TestDeleteRemoteBranchWithoutTheBranchIsFine(t *testing.T) {
	repo := oneRepoWithOrigin(t)
	if err := DeleteRemoteBranch(repo, "feat/first"); err != nil {
		t.Fatalf("a branch origin does not have should be left alone, got %v", err)
	}
}
