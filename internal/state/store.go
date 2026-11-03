package state

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Load reads the state file at path, or returns an empty State when the file does not exist yet.
// It returns an error when the file cannot be read or is not valid JSON.
func Load(path string) (State, error) {
	loaded := State{Sections: map[string]Section{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return loaded, nil
	}
	if err != nil {
		return loaded, err
	}
	if err := json.Unmarshal(data, &loaded); err != nil {
		return loaded, err
	}
	return loaded, nil
}

// Save writes current to path through a temp file and a rename, so a crash never leaves half a file.
// It returns an error when the folder, the temp file or the rename fails.
func Save(path string, current State) error {
	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tempFile, err := os.CreateTemp(filepath.Dir(path), "state-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	return os.Rename(tempFile.Name(), path)
}
