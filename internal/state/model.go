package state

import "time"

// State is .nightcap/state.json: what happened to each heading nightcap took, keyed by heading slug.
type State struct {
	Sections map[string]Section `json:"sections"`
}

// Section is one heading's PR as gh last reported it and how each of its units ended, keyed by unit title.
// PRURL and PRState are a record for people; run decides from gh on every run, never from these.
type Section struct {
	PRURL   string          `json:"pr_url"`
	PRState string          `json:"pr_state"` // OPEN, MERGED or CLOSED; empty when there is no PR
	Units   map[string]Unit `json:"units"`
}

// Unit is how one unit ended after its last round.
type Unit struct {
	Attempts  int       `json:"attempts"`
	Result    string    `json:"result"` // green, read-code, red, blocked or interrupted
	Reason    string    `json:"reason"`
	CheckTail string    `json:"check_tail"`
	UpdatedAt time.Time `json:"updated_at"`
}
