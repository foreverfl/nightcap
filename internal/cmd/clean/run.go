package clean

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"nightcap/internal/client/git"
	"nightcap/internal/store/path"
)

// Run removes, whichever repo they belong to, every worktree folder nightcap left under its cache with the branch checked out in it,
// and every state.json under its data folder, printing each one. Logs and run files stay.
// It prints "nothing to clean" when there is none, and fails when git or the file system cannot remove one.
func Run(stdout io.Writer) error {
	removedWorktrees, err := removeWorktrees(stdout)
	if err != nil {
		return err
	}
	removedStateFiles, err := removeStateFiles(stdout)
	if err != nil {
		return err
	}
	if removedWorktrees == 0 && removedStateFiles == 0 {
		fmt.Fprintln(stdout, "nothing to clean")
	}
	return nil
}

// removeWorktrees removes every folder under ~/.cache/nightcap/*/worktrees and its branch, and returns how many it removed.
func removeWorktrees(stdout io.Writer) (int, error) {
	everyRepoCacheDir, err := path.CacheDir("*")
	if err != nil {
		return 0, err
	}
	worktreeDirs, err := filepath.Glob(filepath.Join(everyRepoCacheDir, "worktrees", "*"))
	if err != nil {
		return 0, err
	}
	for _, worktreeDir := range worktreeDirs {
		repo, err := git.RepoOf(worktreeDir)
		if err != nil {
			return 0, err
		}
		branch, err := git.WorktreeBranch(worktreeDir)
		if err != nil {
			return 0, err
		}
		if err := git.RemoveWorktree(repo, worktreeDir); err != nil {
			return 0, err
		}
		if err := git.DeleteBranch(repo, branch); err != nil {
			return 0, err
		}
		fmt.Fprintf(stdout, "removed %s (%s)\n", worktreeDir, branch)
	}
	return len(worktreeDirs), nil
}

// removeStateFiles removes every ~/.local/share/nightcap/*/state.json without reading it, and returns how many it removed.
func removeStateFiles(stdout io.Writer) (int, error) {
	everyRepoDataDir, err := path.DataDir("*")
	if err != nil {
		return 0, err
	}
	stateFiles, err := filepath.Glob(filepath.Join(everyRepoDataDir, "state.json"))
	if err != nil {
		return 0, err
	}
	for _, stateFile := range stateFiles {
		if err := os.Remove(stateFile); err != nil {
			return 0, err
		}
		fmt.Fprintln(stdout, "removed", stateFile)
	}
	return len(stateFiles), nil
}
