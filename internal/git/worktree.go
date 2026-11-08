package git

import (
	"os"
	"strings"
)

// AddWorktree checks branch out in worktreeDir and returns that folder.
// A new branch starts from base; an existing folder or branch is reused as it is.
func AddWorktree(repo, base, branch, worktreeDir string) (string, error) {
	if _, err := os.Stat(worktreeDir); err == nil {
		return worktreeDir, nil
	}
	args := []string{"worktree", "add", worktreeDir, branch}
	if !branchExists(repo, branch) {
		args = []string{"worktree", "add", "-b", branch, worktreeDir, base}
	}
	if _, err := runGit(repo, args...); err != nil {
		return "", err
	}
	return worktreeDir, nil
}

// RemoveWorktree deletes the worktree at worktreeDir, uncommitted changes included; the branch stays.
func RemoveWorktree(repo, worktreeDir string) error {
	_, err := runGit(repo, "worktree", "remove", "--force", worktreeDir)
	return err
}

// ListWorktrees returns one line per worktree of repo, as `git worktree list` prints them.
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
