package initcmd

import (
	"fmt"
	"io"

	"nightcap/internal/config"
)

// Run writes example as the config file, or says the file already exists and leaves it alone.
func Run(example []byte, stdout io.Writer) error {
	configPath, created, err := config.Init(example)
	if err != nil {
		return err
	}
	if !created {
		fmt.Fprintf(stdout, "%s already exists\n", configPath)
		return nil
	}
	fmt.Fprintf(stdout, "wrote %s, edit it with nightcap config\n", configPath)
	return nil
}
