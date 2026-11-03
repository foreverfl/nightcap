package state

import "time"

// State is .nightcap/state.json: what happened to each heading nightcap took, keyed by heading slug.
type State struct {
	Sections map[string]Section `json:"sections"`
}

// Section is how each unit of one heading ended, keyed by unit slug.
type Section struct {
	Units map[string]Unit `json:"units"`
}

// Unit is how one unit ended after its last round.
type Unit struct {
	Attempts  int       `json:"attempts"`
	Result    string    `json:"result"` // green, read-code, red, blocked or interrupted
	Reason    string    `json:"reason"`
	CheckTail string    `json:"check_tail"`
	UpdatedAt time.Time `json:"updated_at"`
}
