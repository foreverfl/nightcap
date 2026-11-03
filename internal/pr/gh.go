package pr

import "os"

// ghBinary returns the gh to run: $NIGHTCAP_GH when set, so tests can swap in a fake one, otherwise "gh".
func ghBinary() string {
	// Test only: tests set NIGHTCAP_GH to a fake script; a real run leaves it empty.
	if binary := os.Getenv("NIGHTCAP_GH"); binary != "" {
		return binary
	}
	return "gh"
}
