package pr

// autoMergeArgs returns the gh arguments that turn auto-merge on for branch's PR, by rebase when method is "rebase".
func autoMergeArgs(branch, method string) []string {
	flag := "--merge"
	if method == "rebase" {
		flag = "--rebase"
	}
	return []string{"pr", "merge", "--auto", flag, branch}
}

// EnableAutoMerge lets GitHub merge the PR whose head is branch once its required checks pass.
func EnableAutoMerge(worktreeDir, appToken, branch, method string) error {
	_, err := runGhAs(worktreeDir, appToken, autoMergeArgs(branch, method)...)
	return err
}
