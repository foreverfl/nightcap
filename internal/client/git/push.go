package git

// Push pushes branch to origin and sets it as the upstream.
func Push(worktreeDir, branch string) error {
	_, err := runGit(worktreeDir, "push", "-u", "origin", branch)
	return err
}

// DeleteRemoteBranch deletes branch on origin; a branch origin does not have is left alone.
func DeleteRemoteBranch(repo, branch string) error {
	remoteHeads, err := runGit(repo, "ls-remote", "--heads", "origin", "refs/heads/"+branch)
	if err != nil {
		return err
	}
	if remoteHeads == "" {
		return nil
	}
	_, err = runGit(repo, "push", "origin", "--delete", branch)
	return err
}
