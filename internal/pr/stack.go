package pr

// stackArgs returns the gh arguments that tie stackURLs, bottom PR first and the new PR last, into one stack.
func stackArgs(stackURLs []string) []string {
	return append([]string{"stack", "link"}, stackURLs...)
}

// LinkStack ties the PRs at stackURLs, bottom first and the new PR last, into one GitHub stacked PR.
// gh refuses a list that leaves out a PR already in the stack, so every PR below the new one is passed.
func LinkStack(worktreeDir, appToken string, stackURLs []string) error {
	_, err := runGhAs(worktreeDir, appToken, stackArgs(stackURLs)...)
	return err
}
