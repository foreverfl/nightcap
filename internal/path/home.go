package path

import (
	"os"
	"path/filepath"
)

// ConfigFile returns $XDG_CONFIG_HOME/nightcap/nightcap.yml, or ~/.config/nightcap/nightcap.yml when it is unset.
// It does not use os.UserConfigDir, which gives ~/Library/Application Support on a Mac.
func ConfigFile() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "nightcap", "nightcap.yml"), nil
}
