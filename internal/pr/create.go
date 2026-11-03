package pr

// draftArgs returns the gh arguments that open a draft PR titled title with the body in bodyPath.
func draftArgs(title, bodyPath string) []string {
	return []string{"pr", "create", "--draft", "--title", title, "--body-file", bodyPath}
}

// CreateDraft opens a draft PR from dir's branch and returns its URL, which gh prints alone on stdout.
// It returns an error when gh fails.
func CreateDraft(dir, title, body string) (string, error) {
	return runGhWithBody(dir, body, func(bodyPath string) []string { return draftArgs(title, bodyPath) })
}
