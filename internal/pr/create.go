package pr

// createArgs returns the gh arguments that open a PR into base, titled title, with the body in bodyPath.
func createArgs(base, title, bodyPath string) []string {
	return []string{"pr", "create", "--base", base, "--title", title, "--body-file", bodyPath}
}

// Create opens a PR from the worktree's branch into base and returns its URL.
// A non-empty appToken opens it as that App; an empty one opens it as your own gh login.
func Create(worktreeDir, appToken, base, title, body string) (string, error) {
	return runGhWithBody(worktreeDir, appToken, body, func(bodyPath string) []string { return createArgs(base, title, bodyPath) })
}
