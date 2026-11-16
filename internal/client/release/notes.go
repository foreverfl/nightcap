package release

import (
	"regexp"
	"strings"
)

// trailingCommitHash is the " (0b5ac8e…)" release-draft.yml puts after every commit title.
var trailingCommitHash = regexp.MustCompile(` \([0-9a-f]{40}\)$`)

// changeLines returns the "- feat: …", "- fix: …" and "- refactor: …" lines of the releases above version, newest first, without their commit hash.
// It reads every release it was given when version is not among them.
func changeLines(releases []release, version string) []string {
	var lines []string
	for _, entry := range releases {
		if entry.TagName == version {
			break
		}
		for line := range strings.SplitSeq(entry.Body, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "- feat:") || strings.HasPrefix(line, "- fix:") || strings.HasPrefix(line, "- refactor:") {
				lines = append(lines, trailingCommitHash.ReplaceAllString(line, ""))
			}
		}
	}
	return lines
}
