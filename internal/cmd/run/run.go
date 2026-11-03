package run

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"

	"nightcap/internal/check"
	"nightcap/internal/client/claude"
	"nightcap/internal/config"
	"nightcap/internal/git"
	"nightcap/internal/plan"
	"nightcap/internal/pr"
	"nightcap/internal/state"
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
		statePath := filepath.Join(target.Repo, ".nightcap", "state.json")
		saved, err := state.Load(statePath)
		if err != nil {
			return err
		}
		var notBlocked []plan.Section
		hasPR := map[string]bool{}
		for _, section := range plan.Parse(string(planText)) {
			if state.IsBlocked(saved.Sections[section.Slug]) {
				fmt.Fprintf(stdout, "skip %s: blocked\n", section.Title)
				continue
			}
			if len(section.Units) > 0 {
				view, err := pr.ViewBranch(target.Repo, section.Type+"/"+section.Slug)
				if err != nil {
					return err
				}
				if err := state.SavePRStatus(statePath, section.Slug, view.URL, view.State); err != nil {
					return err
				}
				if view.State == "MERGED" || view.State == "CLOSED" {
					fmt.Fprintf(stdout, "skip %s: PR %s\n", section.Title, strings.ToLower(view.State))
					continue
				}
				hasPR[section.Slug] = view.State == "OPEN"
			}
			notBlocked = append(notBlocked, section)
		}
		section, ok := plan.Pick(notBlocked, *sectionName)
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
			if err := runUnit(ctx, stdout, loaded, target, dir, section.Slug, index+1, unit, statePath); err != nil {
				return err
			}
			if err := openDraftPR(ctx, stdout, statePath, dir, section, unit.Title, hasPR[section.Slug], slices.Contains(target.Lang, "en")); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

// runUnit runs claude then the check on one unit up to MaxRounds times, feeding a red check's tail or a red verdict's
// reason into the next prompt. It prints "blocked" when the rounds run out, the same failure comes back twice in a row,
// or claude stops before Done. Every round's outcome goes into the state file at statePath.
// It returns an error only when claude, sh, git or the state file fails.
func runUnit(ctx context.Context, stdout io.Writer, loaded config.Config, target config.Target, dir, slug string, number int, unit plan.Unit, statePath string) error {
	failureTail := ""
	record := state.Unit{}
	for round := 1; round <= loaded.MaxRounds; round++ {
		record.Attempts = round
		logPath := filepath.Join(target.Repo, ".nightcap", "logs", slug, fmt.Sprintf("unit%d-round%d.log", number, round))
		result, err := claude.RunRound(ctx, dir, BuildPrompt(unit.Body, failureTail), loaded.MaxTurns, loaded.AllowedTools, logPath)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "  ran  %s: %s (log: %s)\n", unit.Title, result.Outcome, logPath)
		if result.Outcome != claude.Done {
			break
		}
		// 1. Check: run the check command from nightcap.yml (tests pass or not)
		checked, err := check.Run(ctx, dir, target.Check)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "  check: %s\n", color(checked.Green))
		failure := checked.Tail
		record.Result, record.Reason, record.CheckTail = "red", "", checked.Tail
		// 2. Classify: sort what the branch changed by the rules in check/classify.go
		if checked.Green {
			verdict, err := classifyBranch(stdout, dir, target, unit)
			if err != nil {
				return err
			}
			record.Reason = verdict.Reason
			if !verdict.Red {
				record.Result = "green"
				if verdict.ReadCode {
					record.Result = "read-code"
				}
				return state.SaveUnit(statePath, slug, unit.Title, record)
			}
			failure = verdict.Reason
		}
		if err := state.SaveUnit(statePath, slug, unit.Title, record); err != nil {
			return err
		}
		sameFailureAsLastRound := round > 1 && failure == failureTail
		if sameFailureAsLastRound {
			break
		}
		failureTail = failure
	}
	fmt.Fprintln(stdout, "  blocked")
	record.Result = "blocked"
	return state.SaveUnit(statePath, slug, unit.Title, record)
}

// color names a check result the way the plan file does: "green" for pass, "red" for fail.
func color(green bool) string {
	if green {
		return "green"
	}
	return "red"
}

// classifyBranch prints the files and the count of lines dir's branch changed, then the verdict on them.
// It returns that verdict, or an error when git fails.
func classifyBranch(stdout io.Writer, dir string, target config.Target, unit plan.Unit) (check.Verdict, error) {
	changes, err := git.ChangedFiles(dir)
	if err != nil {
		return check.Verdict{}, err
	}
	for _, change := range changes {
		fmt.Fprintf(stdout, "  %s %s\n", change.Status, change.Path)
	}
	addedLines, err := git.AddedLines(dir)
	if err != nil {
		return check.Verdict{}, err
	}
	fmt.Fprintf(stdout, "  added lines: %d\n", len(addedLines))
	verdict := check.Classify(changes, addedLines, check.WherePaths(unit.Body), target.ReadCode)
	if verdict.Red {
		fmt.Fprintf(stdout, "  verdict: red (%s)\n", verdict.Reason)
	} else if verdict.ReadCode {
		fmt.Fprintf(stdout, "  verdict: green, read the code (%s)\n", verdict.Reason)
	} else {
		fmt.Fprintln(stdout, "  verdict: green")
	}
	return verdict, nil
}

// openDraftPR pushes section's branch from dir once unitTitle has ended green or read-code in the state file at statePath,
// then opens a draft PR, or with hasPR replaces its description, listing every finished unit (English above Korean when
// isEnglishWanted). It returns an error when git, gh or the state file fails.
func openDraftPR(ctx context.Context, stdout io.Writer, statePath, dir string, section plan.Section, unitTitle string, hasPR, isEnglishWanted bool) error {
	saved, err := state.Load(statePath)
	if err != nil {
		return err
	}
	result := saved.Sections[section.Slug].Units[unitTitle].Result
	isAccepted := result == "green" || result == "read-code"
	if !isAccepted {
		return nil
	}
	branch := section.Type + "/" + section.Slug
	if err := git.Push(dir, branch); err != nil {
		return err
	}
	changes, err := git.ChangedFiles(dir)
	if err != nil {
		return err
	}
	addedLines, err := git.AddedLines(dir)
	if err != nil {
		return err
	}
	input := bodyInput(section, saved.Sections[section.Slug], unitTitle)
	input.Risks = pr.RiskLines(changes, addedLines)
	var body string
	if isEnglishWanted {
		translate := func(korean string) (string, error) { return claude.Translate(ctx, korean) }
		body, err = pr.BuildBilingualBody(input, translate)
	} else {
		body, err = pr.BuildBody(input)
	}
	if err != nil {
		return err
	}
	if hasPR {
		return pr.EditBody(dir, branch, body)
	}
	prURL, err := pr.CreateDraft(dir, section.Type+": "+section.Title, body)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "  draft PR: %s\n", prURL)
	return state.SavePRStatus(statePath, section.Slug, prURL, "OPEN")
}

// bodyInput collects the PR description's values: section's summary, every unit of section that ended green or read-code
// in sectionState, in plan order, and the last check tail of lastUnitTitle.
func bodyInput(section plan.Section, sectionState state.Section, lastUnitTitle string) pr.BodyInput {
	input := pr.BodyInput{Summary: section.Summary, CheckTail: sectionState.Units[lastUnitTitle].CheckTail}
	for _, unit := range section.Units {
		record := sectionState.Units[unit.Title]
		if record.Result != "green" && record.Result != "read-code" {
			continue
		}
		bodyUnit := pr.BodyUnit{What: plan.LabelValue(unit.Body, "WHAT"), Done: plan.LabelValue(unit.Body, "DONE")}
		if record.Result == "read-code" {
			bodyUnit.ReadCodeReason = record.Reason
		}
		input.Units = append(input.Units, bodyUnit)
	}
	return input
}
