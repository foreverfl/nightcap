package check

import "strings"

// WherePaths returns the paths on a unit body's "- WHERE" line, split on "," "·" "|",
// with parenthesized notes and backticks dropped. It returns nil when the body has no WHERE line.
func WherePaths(body string) []string {
	for line := range strings.SplitSeq(body, "\n") {
		value, found := strings.CutPrefix(strings.TrimSpace(line), "- WHERE")
		if !found {
			continue
		}
		isSeparator := func(r rune) bool { return r == ',' || r == '·' || r == '|' }
		var paths []string
		for _, field := range strings.FieldsFunc(dropNotes(value), isSeparator) {
			if path := strings.TrimSpace(field); path != "" {
				paths = append(paths, path)
			}
		}
		return paths
	}
	return nil
}

// dropNotes removes backticks and everything inside parentheses, nested ones included, from text.
func dropNotes(text string) string {
	var kept strings.Builder
	depth := 0
	for _, r := range text {
		switch {
		case r == '(':
			depth++
		case r == ')' && depth > 0:
			depth--
		case depth == 0 && r != '`':
			kept.WriteRune(r)
		}
	}
	return kept.String()
}
