package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Event is one line of a run file under .nightcap/runs: something that happened during that nightcap run.
type Event struct {
	Kind      string    `json:"kind"` // run_start, run_end, commit_start, round, test, verdict, blocked or pr_open
	At        time.Time `json:"at"`
	PR        string    `json:"pr,omitempty"`         // the heading's slug: one heading is one PR
	Commit    string    `json:"commit,omitempty"`     // the unit's title: one unit is one commit
	Round     int       `json:"round,omitempty"`      // which round of the unit, from 1
	Seconds   int       `json:"seconds,omitempty"`    // how long claude took on a round
	SessionID string    `json:"session_id,omitempty"` // claude's session, on a round
	Result    string    `json:"result,omitempty"`     // a test's green or red; a verdict's green, read-code or red
	Reason    string    `json:"reason,omitempty"`     // why a run ended early, or why a verdict went red
	URL       string    `json:"url,omitempty"`        // the PR a pr_open line opened
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
