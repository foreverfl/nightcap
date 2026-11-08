package claude

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Outcome is how one claude round ended, before the test runs.
type Outcome string

const (
	Done        Outcome = "done"
	Blocked     Outcome = "blocked"
	RateLimited Outcome = "rate_limited"
)

// Result is one claude round's outcome and its last message.
type Result struct {
	Outcome   Outcome
	Text      string
	SessionID string // pass to the next RunRound to continue this conversation
}

// ParseResult reads the JSON that `claude -p --output-format json` prints.
func ParseResult(raw []byte) (Result, error) {
	var output struct {
		Type           string `json:"type"`
		Result         string `json:"result"`
		SessionID      string `json:"session_id"`
		APIErrorStatus *int   `json:"api_error_status"`
	}
	if err := json.Unmarshal(raw, &output); err != nil || output.Type != "result" {
		return Result{}, fmt.Errorf("claude output is not a result JSON: %.200s", raw)
	}
	result := Result{Outcome: Done, Text: output.Result, SessionID: output.SessionID}
	if output.APIErrorStatus != nil && *output.APIErrorStatus == 429 {
		result.Outcome = RateLimited
	} else if strings.HasPrefix(output.Result, "blocked: needs") {
		result.Outcome = Blocked
	}
	return result, nil
}
