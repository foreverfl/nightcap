package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"nightcap/internal/store/path"
)

// Init writes example to the config file and returns its path; created is false when the file was already there and is left as it is.
func Init(example []byte) (configPath string, created bool, err error) {
	configPath, err = path.ConfigFile()
	if err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return "", false, err
	}
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return configPath, false, nil
	}
	if err != nil {
		return "", false, err
	}
	if _, err := file.Write(example); err != nil {
		file.Close()
		return "", false, err
	}
	return configPath, true, file.Close()
}
