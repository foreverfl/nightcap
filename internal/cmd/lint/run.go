package lint

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"nightcap/internal/store/plan"
	"nightcap/internal/verdict"
)

// Run prints what in the plan file args[0] a night run cannot use: 👤 PRs, PRs without a mark, commits without a test and unreadable WHERE paths.
// It fails when it finds any of those or cannot read the file.
func Run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("lint: name a plan file")
	}
	planText, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	planPRs := plan.Parse(string(planText))
	if len(planPRs) == 0 {
		fmt.Fprintln(stdout, "no PR with type: slug")
		return errors.New("lint: the plan has problems")
	}
	hasProblem := false
	for _, planPR := range planPRs {
		if planPR.IsPerson() {
			fmt.Fprintf(stdout, "person: %s: %s (%s)\n", planPR.Type, planPR.Title, planPR.Slug)
			continue
		}
		if planPR.Mark == plan.MarkNone {
			fmt.Fprintf(stdout, "no mark: %s: %s (%s)\n", planPR.Type, planPR.Title, planPR.Slug)
			hasProblem = true
			continue
		}
		fmt.Fprintf(stdout, "PR: %s: %s (%s)\n", planPR.Type, planPR.Title, planPR.Slug)
		for _, commit := range planPR.Commits {
			if !commit.HasTest {
				fmt.Fprintf(stdout, "  no test: %s\n", commit.Title)
				hasProblem = true
			}
			for _, path := range verdict.WherePaths(commit.Body) {
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
