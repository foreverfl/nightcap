package state

import "time"

// State is .nightcap/state.json: what happened to each PR nightcap took, keyed by the PR's slug.
type State struct {
	PRs map[string]PR `json:"prs"`
}

// PR is one PR as gh last reported it and how each of its commits ended, keyed by the commit's title.
// PRURL and PRState are a record for people; run decides from gh on every run, never from these.
type PR struct {
	PRURL   string            `json:"pr_url"`
	PRState string            `json:"pr_state"` // OPEN, MERGED or CLOSED; empty when there is no PR
	Commits map[string]Commit `json:"commits"`
}

// Commit is how one commit ended after its last round.
type Commit struct {
	Attempts     int           `json:"attempts"`
	Result       string        `json:"result"` // green, read-code, red, blocked, interrupted, or stopped (by a usage limit)
	Reason       string        `json:"reason"`
	TestTail     string        `json:"test_tail"`
	UsageAtLimit *UsageAtLimit `json:"usage_at_limit,omitempty"` // set when claude hit a rate limit during this commit
	UpdatedAt    time.Time     `json:"updated_at"`
}

// UsageAtLimit is the /usage percents read right after claude hit a rate limit, to tell in the morning which limit it was.
type UsageAtLimit struct {
	Session   int `json:"session"`
	WeeklyAll int `json:"weekly_all"`
}
