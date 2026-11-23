package run

import (
	"context"
	"fmt"
	"io"
	"slices"
	"time"

	"nightcap/internal/client/claude"
	"nightcap/internal/client/ghapp"
	"nightcap/internal/client/git"
	"nightcap/internal/client/pr"
	"nightcap/internal/store/config"
	"nightcap/internal/store/events"
	"nightcap/internal/store/plan"
	"nightcap/internal/store/state"
	"nightcap/internal/ui/indent"
)

// openPR pushes planPR's branch and opens its PR, or rewrites the body when hasPR. The caller has checked that commitTitle was accepted.
func openPR(ctx context.Context, stdout io.Writer, nightRun nightRunSettings, worktreeDir, base string, stackBelow []string, planPR plan.PR, commitTitle string, hasPR bool) error {
	saved, err := state.Load(nightRun.statePath)
	if err != nil {
		return err
	}
	branch := planPR.Type + "/" + planPR.Slug
	if err := git.Push(worktreeDir, branch); err != nil {
		return err
	}
	changes, err := git.ChangedFiles(worktreeDir, base)
	if err != nil {
		return err
	}
	addedLines, err := git.AddedLines(worktreeDir, base)
	if err != nil {
		return err
	}
	input := bodyInput(planPR, saved.PRs[planPR.Slug], commitTitle)
	input.Risks = pr.RiskLines(changes, addedLines)
	diff, err := git.Diff(worktreeDir, base)
	if err != nil {
		return err
	}
	reviewOrder, reviewOrderErr := claude.ReviewOrder(ctx, diff)
	input.ReviewOrder, input.IsMissingReviewOrder = reviewOrder, reviewOrderErr != nil
	var body string
	if slices.Contains(nightRun.loaded.Lang, "en") {
		translate := func(korean string) (string, error) { return claude.Translate(ctx, korean) }
		body, err = pr.BuildBilingualBody(input, translate)
	} else {
		body, err = pr.BuildBody(input)
	}
	if err != nil {
		return err
	}
	appToken, err := appToken(ctx, nightRun.loaded, worktreeDir)
	if err != nil {
		return err
	}
	if hasPR {
		return pr.EditBody(worktreeDir, appToken, branch, body)
	}
	title := planPR.Title
	if slices.Contains(nightRun.loaded.Lang, "en") {
		// a failed translation keeps the Korean title, so a title never stops the night
		if english, err := claude.TranslateTitle(ctx, planPR.Title); err == nil && english != "" {
			title = english
		}
	}
	prURL, err := pr.Create(worktreeDir, appToken, base, planPR.Type+": "+title, body)
	if err != nil {
		return err
	}
	indent.Fprintf(stdout, 1, "PR: %s\n", prURL)
	appendEvent(nightRun.runPath, events.Event{Kind: "pr_open", At: time.Now(), PR: planPR.Slug, Commit: commitTitle, URL: prURL})
	if len(stackBelow) > 0 {
		if err := pr.LinkStack(worktreeDir, appToken, append(stackBelow, prURL)); err != nil {
			return err
		}
	}
	return state.SavePRStatus(nightRun.statePath, planPR.Slug, prURL, "OPEN")
}

// appToken returns a fresh token for the App nightcap.yml names on the repo at repoDir, or "" when it names none.
// It is asked right before each use because a token dies after an hour and a night runs longer.
func appToken(ctx context.Context, loaded config.Config, repoDir string) (string, error) {
	if loaded.AppID == 0 || loaded.AppKey == "" {
		return "", nil
	}
	jwt, err := ghapp.JWT(loaded.AppID, loaded.AppKey)
	if err != nil {
		return "", fmt.Errorf("app: %w", err)
	}
	ownerRepo, err := pr.RepoName(repoDir)
	if err != nil {
		return "", err
	}
	token, err := ghapp.Token(ctx, jwt, ownerRepo)
	if err != nil {
		return "", fmt.Errorf("app: %w", err)
	}
	return token, nil
}

// bodyInput returns the PR body's values: planPR's summary and every commit that ended green or read-code.
func bodyInput(planPR plan.PR, savedPR state.PR, lastCommitTitle string) pr.BodyInput {
	input := pr.BodyInput{Summary: planPR.Summary, TestTail: savedPR.Commits[lastCommitTitle].TestTail}
	for _, commit := range planPR.Commits {
		record := savedPR.Commits[commit.Title]
		if !record.IsAccepted() {
			continue
		}
		bodyCommit := pr.BodyCommit{What: plan.LabelValue(commit.Body, "WHAT"), Done: plan.LabelValue(commit.Body, "DONE")}
		if record.Result == "read-code" {
			bodyCommit.ReadCodeReason = record.Reason
		}
		input.Commits = append(input.Commits, bodyCommit)
	}
	return input
}

// labelWhenFinished adds, once every commit is done, the label the PR's mark names, and turns on auto-merge for 🤖.
func labelWhenFinished(ctx context.Context, stdout io.Writer, statePath, worktreeDir string, planPR plan.PR, loaded config.Config) error {
	saved, err := state.Load(statePath)
	if err != nil {
		return err
	}
	for _, commit := range planPR.Commits {
		if !saved.PRs[planPR.Slug].Commits[commit.Title].IsAccepted() {
			return nil
		}
	}
	branch := planPR.Type + "/" + planPR.Slug
	label := pr.ReadingLabel(planPR.Mark)
	appToken, err := appToken(ctx, loaded, worktreeDir)
	if err != nil {
		return err
	}
	if err := pr.AddLabel(worktreeDir, appToken, branch, label); err != nil {
		return err
	}
	indent.Fprintf(stdout, 1, "label: %s\n", label)
	if planPR.Mark == plan.MarkAuto {
		if err := pr.EnableAutoMerge(worktreeDir, appToken, branch, loaded.MergeMethod); err != nil {
			return err
		}
		indent.Fprintf(stdout, 1, "auto-merge: %s\n", branch)
	}
	return nil
}
