package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AddWorktree checks out branch in <repo>/.nightcap/worktrees/<slug> and returns that folder.
// An existing folder is reused as is; an existing branch is checked out instead of created.
// It returns an error when git fails.
func AddWorktree(repo, branch, slug string) (string, error) {
	dir := filepath.Join(repo, ".nightcap", "worktrees", slug)
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}
	args := []string{"worktree", "add", dir, branch}
	if !branchExists(repo, branch) {
		args = []string{"worktree", "add", dir, "-b", branch}
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

// runGit runs git -C repo with args and returns its stdout,
// or an error that carries git's combined output when it exits non-zero.
func runGit(repo string, args ...string) (string, error) {
	command := exec.Command("git", append([]string{"-C", repo}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, output)
	}
	return string(output), nil
}
