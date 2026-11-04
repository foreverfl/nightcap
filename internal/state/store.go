package state

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
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

// SaveUnit records how one unit of the heading sectionSlug ended, stamped with the current time,
// keeping everything else already in the state file at path. It returns an error when reading or writing fails.
func SaveUnit(path, sectionSlug, unitTitle string, unit Unit) error {
	current, err := Load(path)
	if err != nil {
		return err
	}
	section := current.Sections[sectionSlug]
	if section.Units == nil {
		section.Units = map[string]Unit{}
	}
	unit.UpdatedAt = time.Now().UTC()
	section.Units[unitTitle] = unit
	current.Sections[sectionSlug] = section
	return Save(path, current)
}

// IsBlocked reports whether any unit of section ended blocked.
func IsBlocked(section Section) bool {
	for _, unit := range section.Units {
		if unit.Result == "blocked" {
			return true
		}
	}
	return false
}

// SavePRStatus notes the PR URL and state gh reported for the heading sectionSlug, keeping the rest of the state file at path.
// It returns an error when reading or writing fails.
func SavePRStatus(path, sectionSlug, prURL, prState string) error {
	current, err := Load(path)
	if err != nil {
		return err
	}
	section := current.Sections[sectionSlug]
	section.PRURL, section.PRState = prURL, prState
	current.Sections[sectionSlug] = section
	return Save(path, current)
}
