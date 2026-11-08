package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Event is one line of a run file under .nightcap/runs: something that happened during that nightcap run.
type Event struct {
	Kind   string    `json:"kind"` // run_start or run_end
	At     time.Time `json:"at"`
	Reason string    `json:"reason,omitempty"` // why a run ended early; left out when it ran to the end
}

// Append adds event to the file at path as one JSON line, creating the file and its folder when missing.
// It returns an error when the folder, the file or the write fails.
func Append(path string, event Event) error {
	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}
