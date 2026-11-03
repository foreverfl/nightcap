package git

// Push pushes branch from worktreeDir to origin and sets it as the upstream. It returns an error when git fails.
func Push(worktreeDir, branch string) error {
	_, err := runGit(worktreeDir, "push", "-u", "origin", branch)
	return err
}
