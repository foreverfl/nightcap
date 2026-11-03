package pr

// readyArgs returns the gh arguments that turn branch's draft PR into one ready for review.
func readyArgs(branch string) []string {
	return []string{"pr", "ready", branch}
}

// MarkReady turns the draft PR whose head is branch into one ready for review. It returns an error when gh fails.
func MarkReady(dir, branch string) error {
	_, err := runGh(dir, readyArgs(branch)...)
	return err
}
