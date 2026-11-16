package git

// Push pushes branch to origin and sets it as the upstream.
func Push(worktreeDir, branch string) error {
	_, err := runGit(worktreeDir, "push", "-u", "origin", branch)
	return err
}

// DeleteRemoteBranch deletes branch on origin and reports whether there was one to delete; a branch origin does not have is left alone.
func DeleteRemoteBranch(repo, branch string) (bool, error) {
	remoteHeads, err := runGit(repo, "ls-remote", "--heads", "origin", "refs/heads/"+branch)
	if err != nil {
		return false, err
	}
	if remoteHeads == "" {
		return false, nil
	}
	_, err = runGit(repo, "push", "origin", "--delete", branch)
	return err == nil, err
}
