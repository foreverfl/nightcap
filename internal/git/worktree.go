package git

import (
	"os"
	"path/filepath"
	"strings"
)

// AddWorktree checks out branch in <repo>/.nightcap/worktrees/<slug> and returns that folder.
// A new branch starts from base; an existing folder is reused as is and an existing branch is checked out as it is.
// It returns an error when git fails.
func AddWorktree(repo, base, branch, slug string) (string, error) {
	dir := filepath.Join(repo, ".nightcap", "worktrees", slug)
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}
	args := []string{"worktree", "add", dir, branch}
	if !branchExists(repo, branch) {
		args = []string{"worktree", "add", "-b", branch, dir, base}
	}
	if _, err := runGit(repo, args...); err != nil {
		return "", err
	}
	return dir, nil
}

// ListWorktrees returns one line per worktree of repo, as `git worktree list` prints them.
// It returns an error when git fails.
func ListWorktrees(repo string) ([]string, error) {
	output, err := runGit(repo, "worktree", "list")
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(output), "\n"), nil
}

// branchExists reports whether repo has a local branch with that name.
func branchExists(repo, branch string) bool {
	_, err := runGit(repo, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil
}
