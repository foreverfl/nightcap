package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/signal"

	"nightcap/internal/cmd/clean"
	"nightcap/internal/cmd/configcmd"
	"nightcap/internal/cmd/initcmd"
	"nightcap/internal/cmd/lint"
	"nightcap/internal/cmd/run"
	"nightcap/internal/cmd/status"
	"nightcap/internal/cmd/update"
	"nightcap/internal/ui/livescreen"
)

// exampleConfig is examples/nightcap.yml, built into the binary so nightcap init works without the repo.
//
//go:embed examples/nightcap.yml
var exampleConfig []byte

// version is the release tag, set at link time with -ldflags "-X main.version=v0.1.0".
var version = "dev"

const usage = "usage: nightcap run <plan> | nightcap status <plan> | nightcap lint <plan> | nightcap clean | nightcap init | nightcap config | nightcap update | nightcap version\n"

func main() {
	os.Exit(dispatch(os.Args[1:], os.Stdout, os.Stderr))
}

// dispatch runs the subcommand named by args[0] and returns the exit code:
// 0 on success, 1 when the subcommand fails, 2 when the command is missing or unknown.
func dispatch(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	var err error
	switch args[0] {
	case "run":
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if isTerminal(stdout) {
			err = livescreen.Run(ctx, func(ctx context.Context, stdout io.Writer) error {
				return run.Run(ctx, args[1:], version, stdout)
			})
		} else {
			err = run.Run(ctx, args[1:], version, stdout)
		}
	case "status":
		err = status.Run(args[1:], stdout)
	case "lint":
		err = lint.Run(args[1:], stdout)
	case "clean":
		err = clean.Run(stdout)
	case "init":
		err = initcmd.Run(exampleConfig, stdout)
	case "config":
		err = configcmd.Run()
	case "update":
		err = update.Run(stdout, stderr)
	case "version":
		fmt.Fprintln(stdout, "nightcap", version)
	default:
		fmt.Fprint(stderr, usage)
		return 2
	}
	if err != nil {
		fmt.Fprintln(stderr, "nightcap:", err)
		return 1
	}
	return 0
}

// isTerminal reports whether output is a terminal, not a pipe, a file or a test buffer.
func isTerminal(output io.Writer) bool {
	file, ok := output.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
