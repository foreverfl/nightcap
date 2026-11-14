package git

// Push pushes branch to origin and sets it as the upstream.
func Push(worktreeDir, branch string) error {
	_, err := runGit(worktreeDir, "push", "-u", "origin", branch)
	return err
}
