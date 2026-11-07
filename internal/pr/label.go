package pr

import "nightcap/internal/plan"

const (
	autoLabel   = "nightcap:auto"
	readPRLabel = "nightcap:read-pr"
)

// ReadingLabel returns the label the heading's reading mark names: nightcap:auto for 🤖,
// otherwise nightcap:read-pr. What the night run found never changes it; that goes to the risk section instead.
func ReadingLabel(mark plan.Mark) string {
	switch mark {
	case plan.MarkAuto:
		return autoLabel
	}
	return readPRLabel
}

// AddLabel creates label in the repo when it is missing and puts it on the PR whose head is branch.
// It returns an error when gh fails.
func AddLabel(worktreeDir, branch, label string) error {
	if _, err := runGh(worktreeDir, "label", "create", label, "--force"); err != nil {
		return err
	}
	_, err := runGh(worktreeDir, "pr", "edit", branch, "--add-label", label)
	return err
}
