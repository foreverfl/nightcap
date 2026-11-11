package indent

import (
	"strings"
	"testing"
)

func TestFprintfPutsFourSpacesAtDepthTwo(t *testing.T) {
	var output strings.Builder

	Fprintf(&output, 2, "Edit %s\n", "a.go")

	if output.String() != "    Edit a.go\n" {
		t.Fatalf("output = %q", output.String())
	}
}
