package configcmd

import "nightcap/internal/store/config"

// Run opens the config file in the user's editor.
func Run() error {
	return config.Edit()
}
