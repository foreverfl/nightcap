package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"

	"nightcap/internal/cmd/clean"
	"nightcap/internal/cmd/configcmd"
	"nightcap/internal/cmd/initcmd"
	"nightcap/internal/cmd/lint"
	"nightcap/internal/cmd/run"
	"nightcap/internal/cmd/status"
	"nightcap/internal/cmd/update"
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
		err = run.Run(args[1:], stdout)
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
