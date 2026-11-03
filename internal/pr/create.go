package pr

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// draftArgs returns the gh arguments that open a draft PR titled title with the body in bodyPath.
func draftArgs(title, bodyPath string) []string {
	return []string{"pr", "create", "--draft", "--title", title, "--body-file", bodyPath}
}

// CreateDraft opens a draft PR from dir's branch and returns its URL, which gh prints alone on stdout.
// The binary is $NIGHTCAP_GH when set, so tests can swap in a fake one. It returns an error when gh fails.
func CreateDraft(dir, title, body string) (string, error) {
	bodyFile, err := os.CreateTemp("", "nightcap-pr-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(bodyFile.Name())
	_, writeErr := bodyFile.WriteString(body)
	if closeErr := bodyFile.Close(); writeErr != nil || closeErr != nil {
		return "", errors.Join(writeErr, closeErr)
	}
	// Test only: main_test.go sets NIGHTCAP_GH to a fake script; a real run leaves it empty.
	binary := os.Getenv("NIGHTCAP_GH")
	if binary == "" {
		binary = "gh"
	}
	command := exec.Command(binary, draftArgs(title, bodyFile.Name())...)
	command.Dir = dir
	output, err := command.Output()
	if exitErr, isExitErr := errors.AsType[*exec.ExitError](err); isExitErr {
		return "", fmt.Errorf("gh pr create: %w\n%s", err, exitErr.Stderr)
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
