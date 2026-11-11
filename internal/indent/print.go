package indent

import (
	"fmt"
	"io"
	"strings"
)

// Fprintf is fmt.Fprintf with two spaces per depth put before format.
// depth 1 is a step nightcap takes; depth 2 is a step claude takes inside it.
func Fprintf(w io.Writer, depth int, format string, args ...any) {
	fmt.Fprintf(w, strings.Repeat("  ", depth)+format, args...)
}
