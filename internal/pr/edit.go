package pr

// editArgs returns the gh arguments that replace the description of branch's PR with the body in bodyPath.
func editArgs(branch, bodyPath string) []string {
	return []string{"pr", "edit", branch, "--body-file", bodyPath}
}

// EditBody replaces the body of the PR whose head is branch.
func EditBody(worktreeDir, branch, body string) error {
	_, err := runGhWithBody(worktreeDir, body, func(bodyPath string) []string { return editArgs(branch, bodyPath) })
	return err
}
