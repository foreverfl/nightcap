package status

import (
	"fmt"
	"io"

	"nightcap/internal/config"
	"nightcap/internal/git"
)

// Run is the "status" subcommand; args is everything after "status" (unused yet).
// For now it prints each target repo with its worktrees, and returns an error
// when the config does not load or git fails.
func Run(args []string, stdout io.Writer) error {
	loaded, err := config.Load(".")
	if err != nil {
		return err
	}
	for _, target := range loaded.Targets {
		worktrees, err := git.ListWorktrees(target.Repo)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, target.Repo)
		for _, line := range worktrees {
			fmt.Fprintln(stdout, "  "+line)
		}
	}
	return nil
}
