package plan

import "strings"

// LabelValue returns the text after "- <label>" on a unit body's line for that label, such as "WHAT" or "DONE",
// or an empty string when the body has no such line.
func LabelValue(body, label string) string {
	for line := range strings.SplitSeq(body, "\n") {
		if value, found := strings.CutPrefix(strings.TrimSpace(line), "- "+label+" "); found {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
