package pr

import "slices"

const (
	readPRLabel   = "nightcap:read-pr"
	readCodeLabel = "nightcap:read-code"
)

// ReadingLabel returns nightcap:read-code when any of the units' results is read-code, otherwise nightcap:read-pr.
func ReadingLabel(unitResults []string) string {
	if slices.Contains(unitResults, "read-code") {
		return readCodeLabel
	}
	return readPRLabel
}

// AddLabel creates label in the repo when it is missing and puts it on the PR whose head is branch.
// It returns an error when gh fails.
func AddLabel(dir, branch, label string) error {
	if _, err := runGh(dir, "label", "create", label, "--force"); err != nil {
		return err
	}
	_, err := runGh(dir, "pr", "edit", branch, "--add-label", label)
	return err
}
