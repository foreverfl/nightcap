package verdict

import "strings"

// WherePaths returns the paths on a unit body's "- WHERE" line, split on "," "·" "|" outside braces,
// with "dir/{a.go, b.go}" expanded and parenthesized notes and backticks dropped.
// It returns nil when the body has no WHERE line.
func WherePaths(body string) []string {
	for line := range strings.SplitSeq(body, "\n") {
		value, found := strings.CutPrefix(strings.TrimSpace(line), "- WHERE")
		if !found {
			continue
		}
		var paths []string
		for _, field := range splitOutsideBraces(dropNotes(value)) {
			for _, path := range expandBraces(strings.TrimSpace(field)) {
				if path != "" {
					paths = append(paths, path)
				}
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

// splitOutsideBraces splits text on "," "·" "|", leaving the commas inside {…} alone.
func splitOutsideBraces(text string) []string {
	var fields []string
	var current strings.Builder
	braceDepth := 0
	for _, r := range text {
		switch {
		case r == '{':
			braceDepth++
		case r == '}' && braceDepth > 0:
			braceDepth--
		case braceDepth == 0 && (r == ',' || r == '·' || r == '|'):
			fields = append(fields, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	return append(fields, current.String())
}

// expandBraces turns "dir/{a.go, b.go}" into ["dir/a.go", "dir/b.go"]; an item without braces comes back alone.
func expandBraces(item string) []string {
	openBrace, closeBrace := strings.Index(item, "{"), strings.LastIndex(item, "}")
	if openBrace < 0 || closeBrace < openBrace {
		return []string{item}
	}
	prefix, suffix := item[:openBrace], item[closeBrace+1:]
	var paths []string
	for name := range strings.SplitSeq(item[openBrace+1:closeBrace], ",") {
		paths = append(paths, prefix+strings.TrimSpace(name)+suffix)
	}
	return paths
}
