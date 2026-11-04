package claude

import "os"

// claudeBinary returns the claude to run: $NIGHTCAP_CLAUDE when set, so tests can swap in a fake one, otherwise "claude".
func claudeBinary() string {
	// Test only: tests set NIGHTCAP_CLAUDE to a fake script; a real run leaves it empty.
	if binary := os.Getenv("NIGHTCAP_CLAUDE"); binary != "" {
		return binary
	}
	return "claude"
}
