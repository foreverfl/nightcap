package state

import "testing"

func TestIsAcceptedTakesGreenAndReadCodeOnly(t *testing.T) {
	for result, want := range map[string]bool{"green": true, "read-code": true, "red": false} {
		if got := (Commit{Result: result}).IsAccepted(); got != want {
			t.Errorf("Result %q: want %v, got %v", result, want, got)
		}
	}
}
