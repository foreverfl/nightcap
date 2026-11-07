package pr

// createArgs returns the gh arguments that open a PR into base, titled title, with the body in bodyPath.
func createArgs(base, title, bodyPath string) []string {
	return []string{"pr", "create", "--base", base, "--title", title, "--body-file", bodyPath}
}

// Create opens a PR from worktreeDir's branch into base and returns its URL, which gh prints alone on stdout.
// It returns an error when gh fails.
func Create(worktreeDir, base, title, body string) (string, error) {
	return runGhWithBody(worktreeDir, body, func(bodyPath string) []string { return createArgs(base, title, bodyPath) })
}
