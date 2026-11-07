package pr

// autoMergeArgs returns the gh arguments that let GitHub merge branch's PR once the required checks pass,
// by rebase when method is "rebase" and with a merge commit otherwise.
func autoMergeArgs(branch, method string) []string {
	flag := "--merge"
	if method == "rebase" {
		flag = "--rebase"
	}
	return []string{"pr", "merge", "--auto", flag, branch}
}

// EnableAutoMerge lets GitHub merge the PR whose head is branch once its checks pass. It returns an error when gh fails.
func EnableAutoMerge(worktreeDir, branch, method string) error {
	_, err := runGh(worktreeDir, autoMergeArgs(branch, method)...)
	return err
}
