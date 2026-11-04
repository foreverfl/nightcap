package lint

import (
	"errors"
	"fmt"
	"io"
	"os"

	"nightcap/internal/plan"
)

// Run is the "lint" subcommand; args[0] is the plan file. It prints every heading a night run can read
// and every open unit whose DONE names no check, and returns an error when the file has no readable heading,
// a unit has no check, or the file cannot be read.
func Run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("lint: name a plan file")
	}
	planText, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	sections := plan.Parse(string(planText))
	if len(sections) == 0 {
		fmt.Fprintln(stdout, "no heading with type: slug")
		return errors.New("lint: the plan has problems")
	}
	hasProblem := false
	for _, section := range sections {
		fmt.Fprintf(stdout, "heading: %s: %s (%s)\n", section.Type, section.Title, section.Slug)
		for _, unit := range section.Units {
			if !unit.HasCheck {
				fmt.Fprintf(stdout, "  no check: %s\n", unit.Title)
				hasProblem = true
			}
		}
	}
	if hasProblem {
		return errors.New("lint: the plan has problems")
	}
	return nil
}
