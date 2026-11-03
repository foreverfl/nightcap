package git

// Push pushes branch from dir to origin and sets it as the upstream. It returns an error when git fails.
func Push(dir, branch string) error {
	_, err := runGit(dir, "push", "-u", "origin", branch)
	return err
}
