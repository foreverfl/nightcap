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
// one claude round on the first unit that has a check, then the target's check command when that
// round ended Done; it returns an error when the config, a plan file, git, claude or sh fails.
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
			logPath := filepath.Join(target.Repo, ".nightcap", "logs", section.Slug, fmt.Sprintf("unit%d-round1.log", index+1))
			prompt := BuildPrompt(unit.Body, "")
			result, err := claude.RunRound(ctx, dir, prompt, loaded.MaxTurns, loaded.AllowedTools, logPath)
			if err != nil {
				return err
			}
			fmt.Fprintf(stdout, "  ran  %s: %s (log: %s)\n", unit.Title, result.Outcome, logPath)
			if result.Outcome == claude.Done {
				checked, err := check.Run(ctx, dir, target.Check)
				if err != nil {
					return err
				}
				fmt.Fprintf(stdout, "  check: %s\n", color(checked.Green))
			}
			break
		}
	}
	return nil
}

// color names a check result the way the plan file does: "green" for pass, "red" for fail.
func color(green bool) string {
	if green {
		return "green"
	}
	return "red"
}
