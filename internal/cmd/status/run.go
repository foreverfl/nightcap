package status

import (
	"errors"
	"fmt"
	"io"
)

// Run is the "status" subcommand; args is everything after "status".
// For now it only prints the target repo path, and returns an error when the path is missing.
func Run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("status: repo path is required")
	}
	fmt.Fprintln(stdout, "repo:", args[0])
	return nil
}
