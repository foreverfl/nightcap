package pr

// stackArgs returns the gh arguments that put newURL on top of the stack previousURL sits in, or start one with both.
func stackArgs(previousURL, newURL string) []string {
	return []string{"stack", "link", previousURL, newURL}
}

// LinkStack ties the PR at newURL onto the PR at previousURL as a GitHub stacked PR.
func LinkStack(worktreeDir, appToken, previousURL, newURL string) error {
	_, err := runGhAs(worktreeDir, appToken, stackArgs(previousURL, newURL)...)
	return err
}
