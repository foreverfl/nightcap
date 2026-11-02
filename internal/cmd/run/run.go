package run

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"

	"nightcap/internal/check"
	"nightcap/internal/client/claude"
	"nightcap/internal/config"
	"nightcap/internal/git"
	"nightcap/internal/plan"
)

// Run is the "run" subcommand; args is everything after "run", where --section names one heading.
// For now it picks tonight's heading per target, checks its branch out in a worktree and runs
// the first unit that has a check through runUnit; it returns an error when the config,
// a plan file, git, claude or sh fails.
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

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

		for index, unit := range section.Units {
			if !unit.HasCheck {
				continue
			}
			if err := runUnit(ctx, stdout, loaded, target, dir, section.Slug, index+1, unit); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

// runUnit runs claude and then the check on one unit, up to MaxRounds times, feeding the failing
// check's tail into the next prompt. It prints one "ran" and one "check" line per round, and
// "blocked" when the rounds run out or claude stops before Done; it returns an error when
// claude or sh cannot run.
func runUnit(ctx context.Context, stdout io.Writer, loaded config.Config, target config.Target, dir, slug string, number int, unit plan.Unit) error {
	failureTail := ""
	for round := 1; round <= loaded.MaxRounds; round++ {
		logPath := filepath.Join(target.Repo, ".nightcap", "logs", slug, fmt.Sprintf("unit%d-round%d.log", number, round))
		result, err := claude.RunRound(ctx, dir, BuildPrompt(unit.Body, failureTail), loaded.MaxTurns, loaded.AllowedTools, logPath)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "  ran  %s: %s (log: %s)\n", unit.Title, result.Outcome, logPath)
		if result.Outcome != claude.Done {
			break
		}
		checked, err := check.Run(ctx, dir, target.Check)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "  check: %s\n", color(checked.Green))
		if checked.Green {
			return nil
		}
		failureTail = checked.Tail
	}
	fmt.Fprintln(stdout, "  blocked")
	return nil
}

// color names a check result the way the plan file does: "green" for pass, "red" for fail.
func color(green bool) string {
	if green {
		return "green"
	}
	return "red"
}
