package run

import (
	"flag"
	"fmt"
	"io"
	"os"

	"nightcap/internal/config"
	"nightcap/internal/git"
	"nightcap/internal/plan"
)

// Run is the "run" subcommand; args is everything after "run", where --section names one heading.
// For now it prints, per target, which heading and units tonight's run would take and
// checks that heading's branch out in a worktree; it returns an error when the config,
// a plan file or git fails.
func Run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	sectionName := flags.String("section", "", "heading title or slug to take")
	if err := flags.Parse(args); err != nil {
		return err
	}
	loaded, err := config.Load(".")
	if err != nil {
		return err
	}
	for _, target := range loaded.Targets {
		planText, err := os.ReadFile(target.Plan)
		if err != nil {
			return err
		}
		section, ok := plan.Pick(plan.Parse(string(planText)), *sectionName)
		if !ok {
			fmt.Fprintf(stdout, "%s: nothing to take\n", target.Plan)
			continue
		}
		branch := section.Type + "/" + section.Slug
		dir, err := git.AddWorktree(target.Repo, branch, section.Slug)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "%s\n  branch: %s\n  worktree: %s\n", target.Plan, branch, dir)
		for _, unit := range section.Units {
			mark := "skip"
			if unit.HasCheck {
				mark = "take"
			}
			fmt.Fprintf(stdout, "  %s  %s\n", mark, unit.Title)
		}
	}
	return nil
}
