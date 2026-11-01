package run

import (
	"flag"
	"fmt"
	"io"
	"os"

	"nightcap/internal/config"
	"nightcap/internal/plan"
)

// Run is the "run" subcommand; args is everything after "run", where --section names one heading.
// For now it only prints, per target, which heading and which units tonight's run would take,
// and returns an error when the config or a plan file does not load.
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
		fmt.Fprintf(stdout, "%s\n  branch: %s/%s\n", target.Plan, section.Type, section.Slug)
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
