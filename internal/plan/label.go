package plan

import "strings"

// LabelValue returns the text after "- <label>" in body, such as its WHAT or DONE line, or "" when there is none.
func LabelValue(body, label string) string {
	for line := range strings.SplitSeq(body, "\n") {
		if value, found := strings.CutPrefix(strings.TrimSpace(line), "- "+label+" "); found {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
