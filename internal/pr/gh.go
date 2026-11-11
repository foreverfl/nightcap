package pr

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ghBinary returns the gh to run: $NIGHTCAP_GH when set, so tests can swap in a fake one, otherwise "gh".
func ghBinary() string {
	// Test only: tests set NIGHTCAP_GH to a fake script; a real run leaves it empty.
	if binary := os.Getenv("NIGHTCAP_GH"); binary != "" {
		return binary
	}
	return "gh"
}

// runGhWithBody runs gh as appToken's App, or as your own gh login when it is empty, with the arguments argsFor builds around a temp file holding body.
// It returns gh's stdout.
func runGhWithBody(worktreeDir, appToken, body string, argsFor func(bodyPath string) []string) (string, error) {
	bodyFile, err := os.CreateTemp("", "nightcap-pr-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(bodyFile.Name())
	_, writeErr := bodyFile.WriteString(body)
	if closeErr := bodyFile.Close(); writeErr != nil || closeErr != nil {
		return "", errors.Join(writeErr, closeErr)
	}
	return runGhAs(worktreeDir, appToken, argsFor(bodyFile.Name())...)
}

// runGh runs gh in worktreeDir with args as your own gh login and returns its trimmed stdout; a failure carries gh's stderr.
func runGh(worktreeDir string, args ...string) (string, error) {
	return runGhAs(worktreeDir, "", args...)
}

// runGhAs is runGh run as appToken's App when appToken is not empty.
func runGhAs(worktreeDir, appToken string, args ...string) (string, error) {
	command := exec.Command(ghBinary(), args...)
	command.Dir = worktreeDir
	if appToken != "" {
		command.Env = append(os.Environ(), "GH_TOKEN="+appToken)
	}
	output, err := command.Output()
	if exitErr, isExitErr := errors.AsType[*exec.ExitError](err); isExitErr {
		return "", fmt.Errorf("gh %s: %w\n%s", strings.Join(args[:2], " "), err, exitErr.Stderr)
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
