package pr

// autoMergeArgs returns the gh arguments that let GitHub merge branch's PR once the required checks pass.
func autoMergeArgs(branch string) []string {
	return []string{"pr", "merge", "--auto", "--merge", branch}
}

// EnableAutoMerge lets GitHub merge the PR whose head is branch once its checks pass. It returns an error when gh fails.
func EnableAutoMerge(worktreeDir, branch string) error {
	_, err := runGh(worktreeDir, autoMergeArgs(branch)...)
	return err
}
