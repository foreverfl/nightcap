package pr

// RepoName returns the GitHub owner/repo of the repo at repoDir, e.g. "foreverfl/blog-v2", asked as your own gh login.
func RepoName(repoDir string) (string, error) {
	return runGh(repoDir, "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
}
