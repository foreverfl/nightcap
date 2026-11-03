package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// runGit runs git -C repo with args and returns its stdout,
// or an error that carries git's combined output when it exits non-zero.
func runGit(repo string, args ...string) (string, error) {
	command := exec.Command("git", append([]string{"-C", repo}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, output)
	}
	return string(output), nil
}
