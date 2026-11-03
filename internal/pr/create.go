package pr

// draftArgs returns the gh arguments that open a draft PR into base, titled title, with the body in bodyPath.
func draftArgs(base, title, bodyPath string) []string {
	return []string{"pr", "create", "--draft", "--base", base, "--title", title, "--body-file", bodyPath}
}

// CreateDraft opens a draft PR from dir's branch into base and returns its URL, which gh prints alone on stdout.
// It returns an error when gh fails.
func CreateDraft(dir, base, title, body string) (string, error) {
	return runGhWithBody(dir, body, func(bodyPath string) []string { return draftArgs(base, title, bodyPath) })
}
