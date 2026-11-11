package update

import (
	"io"
	"os/exec"
)

const installCommand = "curl -fsSL https://raw.githubusercontent.com/foreverfl/nightcap/main/install.sh | sh"

// Run downloads install.sh from main and runs it, so the latest release replaces ~/.local/bin/nightcap.
// install.sh's output goes straight to stdout and stderr; it fails when curl or install.sh exits non-zero.
func Run(stdout, stderr io.Writer) error {
	command := exec.Command("sh", "-c", installCommand)
	command.Stdout = stdout
	command.Stderr = stderr
	return command.Run()
}
