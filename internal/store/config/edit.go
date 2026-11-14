package config

import (
	"fmt"
	"os"
	"os/exec"

	"nightcap/internal/store/path"
)

// Edit opens the config file in $VISUAL, then $EDITOR, then vi, and waits for the editor to close.
// It fails when the config file does not exist yet or the editor exits with an error.
func Edit() error {
	configPath, err := path.ConfigFile()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("config: %s not found, run nightcap init", configPath)
	}
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	// through sh, as git does, so an editor with its own arguments like "zed --wait" works
	command := exec.Command("sh", "-c", editor+` "$1"`, "sh", configPath)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	return command.Run()
}
