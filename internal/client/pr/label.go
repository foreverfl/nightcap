package pr

import "nightcap/internal/store/plan"

const (
	autoLabel   = "nightcap:auto"
	readPRLabel = "nightcap:read-pr"

	autoLabelColor   = "0E8A16"
	readPRLabelColor = "7057FF"
)

// labelColor returns the hex color, without "#", that label is created with: green for nightcap:auto, purple otherwise.
func labelColor(label string) string {
	if label == autoLabel {
		return autoLabelColor
	}
	return readPRLabelColor
}

// ReadingLabel returns the label the PR's reading mark names: nightcap:auto for 🤖,
// otherwise nightcap:read-pr. What the night run found never changes it; that goes to the risk section instead.
func ReadingLabel(mark plan.Mark) string {
	switch mark {
	case plan.MarkAuto:
		return autoLabel
	}
	return readPRLabel
}

// AddLabel puts label on the PR whose head is branch, creating the label in the repo when it is missing.
func AddLabel(worktreeDir, appToken, branch, label string) error {
	if _, err := runGhAs(worktreeDir, appToken, "label", "create", label, "--color", labelColor(label), "--force"); err != nil {
		return err
	}
	_, err := runGhAs(worktreeDir, appToken, "pr", "edit", branch, "--add-label", label)
	return err
}
