package clean

import (
	"fmt"
	"io"
	"path/filepath"

	"nightcap/internal/git"
	"nightcap/internal/path"
)

// Run removes every worktree folder nightcap left under its cache, whichever repo it belongs to, and prints each one.
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
		if err := git.RemoveWorktree(repo, worktreeDir); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "removed", worktreeDir)
	}
	return nil
}
