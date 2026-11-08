package configcmd

import "nightcap/internal/config"

// Run opens the config file in the user's editor.
func Run() error {
	return config.Edit()
}
