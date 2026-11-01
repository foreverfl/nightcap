package run

import (
	"fmt"
	"io"

	"nightcap/internal/config"
)

// Run is the "run" subcommand; args is everything after "run" (an optional plan path, unused yet).
// For now it only loads nightcap.yml from the current folder and prints it,
// and returns an error when the config does not load.
func Run(args []string, stdout io.Writer) error {
	loaded, err := config.Load(".")
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "max_turns: %d  max_rounds: %d  work_start: %s\n", loaded.MaxTurns, loaded.MaxRounds, loaded.WorkStart)
	for _, target := range loaded.Targets {
		fmt.Fprintf(stdout, "target: %s -> %s (%s)\n", target.Plan, target.Repo, target.Check)
	}
	return nil
}
