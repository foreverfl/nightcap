package path

import (
	"os"
	"path/filepath"
)

// ConfigFile returns $XDG_CONFIG_HOME/nightcap/nightcap.yml, or ~/.config/nightcap/nightcap.yml when it is unset.
// It does not use os.UserConfigDir, which gives ~/Library/Application Support on a Mac.
func ConfigFile() (string, error) {
	base, err := baseDir("XDG_CONFIG_HOME", ".config")
	return filepath.Join(base, "nightcap", "nightcap.yml"), err
}

// DataDir returns $XDG_DATA_HOME/nightcap/<repo folder name>, or ~/.local/share/nightcap/<repo folder name> when it is unset.
func DataDir(repo string) (string, error) {
	base, err := baseDir("XDG_DATA_HOME", filepath.Join(".local", "share"))
	return filepath.Join(base, "nightcap", filepath.Base(repo)), err
}

// baseDir returns the folder the env variable names, or homeFolder under the home folder when it is unset.
func baseDir(env, homeFolder string) (string, error) {
	if base := os.Getenv(env); base != "" {
		return base, nil
	}
	home, err := os.UserHomeDir()
	return filepath.Join(home, homeFolder), err
}
