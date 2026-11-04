package main

import (
	"fmt"
	"io"
	"os"

	"nightcap/internal/cmd/lint"
	"nightcap/internal/cmd/run"
	"nightcap/internal/cmd/status"
)

const usage = "usage: nightcap run [plan] [--section <title>] | nightcap status [plan] | nightcap lint <plan>\n"

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
