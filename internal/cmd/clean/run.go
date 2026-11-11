package clean

import (
	"fmt"
	"io"
	"path/filepath"

	"nightcap/internal/git"
	"nightcap/internal/path"
)

// Run removes every worktree folder nightcap left under its cache and the branch checked out in it, whichever repo it belongs to, and prints each one.
// It prints "nothing to clean" when there is none, and fails when git cannot find or remove one.
func Run(stdout io.Writer) error {
	everyRepoCacheDir, err := path.CacheDir("*")
	if err != nil {
		return err
	}
	worktreeDirs, err := filepath.Glob(filepath.Join(everyRepoCacheDir, "worktrees", "*"))
	if err != nil {
		return err
	}
	if len(worktreeDirs) == 0 {
		fmt.Fprintln(stdout, "nothing to clean")
		return nil
	}
	for _, worktreeDir := range worktreeDirs {
		repo, err := git.RepoOf(worktreeDir)
		if err != nil {
			return err
		}
		branch, err := git.WorktreeBranch(worktreeDir)
		if err != nil {
			return err
		}
		if err := git.RemoveWorktree(repo, worktreeDir); err != nil {
			return err
		}
		if err := git.DeleteBranch(repo, branch); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "removed %s (%s)\n", worktreeDir, branch)
	}
	return nil
}
