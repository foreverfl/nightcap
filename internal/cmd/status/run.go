package status

import (
	"errors"
	"fmt"
	"io"
	"os"

	"nightcap/internal/git"
	"nightcap/internal/plan"
)

// Run is the "status" subcommand; args is everything after "status", and its first item is the plan file.
// It prints the repo the plan's frontmatter names with its worktrees, and returns an error
// when no plan file is named, the plan file or its frontmatter does not read, or git fails.
func Run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("status: name a plan file")
	}
	planText, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	meta, err := plan.ParseMeta(string(planText))
	if err != nil {
		return err
	}
	if meta.Repo == "" {
		return errors.New("status: the plan's frontmatter needs repo")
	}
	worktrees, err := git.ListWorktrees(meta.Repo)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, meta.Repo)
	for _, line := range worktrees {
		fmt.Fprintln(stdout, "  "+line)
	}
	return nil
}
