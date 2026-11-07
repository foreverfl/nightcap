package lint

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"nightcap/internal/check"
	"nightcap/internal/plan"
)

// Run is the "lint" subcommand; args[0] is the plan file. It prints every heading a night run can read, each 👤 heading as person, each heading without a reading mark as no mark,
// every open unit whose DONE names no check, and every WHERE path the run cannot read inside the repo,
// and returns an error when any of those is found or the file cannot be read.
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
		if section.IsPerson() {
			fmt.Fprintf(stdout, "person: %s: %s (%s)\n", section.Type, section.Title, section.Slug)
			continue
		}
		if section.Mark == plan.MarkNone {
			fmt.Fprintf(stdout, "no mark: %s: %s (%s)\n", section.Type, section.Title, section.Slug)
			hasProblem = true
			continue
		}
		fmt.Fprintf(stdout, "heading: %s: %s (%s)\n", section.Type, section.Title, section.Slug)
		for _, unit := range section.Units {
			if !unit.HasCheck {
				fmt.Fprintf(stdout, "  no check: %s\n", unit.Title)
				hasProblem = true
			}
			for _, path := range check.WherePaths(unit.Body) {
				if canNeverMatchAChangedFile(path) {
					fmt.Fprintf(stdout, "  WHERE path not readable: %s\n", path)
					hasProblem = true
				}
			}
		}
	}
	if hasProblem {
		return errors.New("lint: the plan has problems")
	}
	return nil
}

// canNeverMatchAChangedFile reports a WHERE path no git path can equal, since git names changed files from the repo root.
func canNeverMatchAChangedFile(path string) bool {
	return strings.Contains(path, " ") || strings.HasPrefix(path, "/") || strings.HasPrefix(path, "~") || strings.Contains(path, "..")
}
