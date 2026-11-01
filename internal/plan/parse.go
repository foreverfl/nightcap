package plan

import (
	"bufio"
	"strings"
)

var sectionTypes = []string{"feat", "fix", "test", "docs", "refactor", "chore"}

var checkNames = []string{"hurl", "go test", "make check", "cargo test", "npm test"}

// Parse returns every "type: title (slug)" heading in a plan file with its open units, in file order.
// Headings without a known type or a slug are skipped together with the units under them.
func Parse(text string) []Section {
	var sections []Section
	var current *Unit
	underTypedHeading := false
	inFence := false

	flush := func() {
		if current == nil {
			return
		}
		current.Body = strings.TrimRight(current.Body, "\n")
		current.HasCheck = hasCheck(current.Body)
		last := &sections[len(sections)-1]
		last.Units = append(last.Units, *current)
		current = nil
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		isFence := strings.HasPrefix(strings.TrimSpace(line), "```")
		if inFence {
			if current != nil {
				current.Body += line + "\n"
			}
			inFence = !isFence
			continue
		}
		if isFence {
			inFence = true
		}
		if current != nil && (line == "" || isIndented(line)) {
			current.Body += line + "\n"
			continue
		}
		flush()
		if isFence {
			continue
		}
		if strings.HasPrefix(line, "#") {
			section, ok := parseHeading(line)
			underTypedHeading = ok
			if ok {
				sections = append(sections, section)
			}
			continue
		}
		if !underTypedHeading {
			continue
		}
		if title, ok := strings.CutPrefix(line, "- [ ] "); ok {
			current = &Unit{Title: title, Body: line + "\n"}
			continue
		}
		last := &sections[len(sections)-1]
		isPlainLine := line != "" && !isIndented(line) && !strings.HasPrefix(line, "- ")
		if last.Summary == "" && len(last.Units) == 0 && isPlainLine {
			last.Summary = line
		}
	}
	flush()
	return sections
}

// Pick returns the section whose Title or Slug equals name, or with an empty name
// the first section that has a unit a night run can take; ok is false when none matches.
func Pick(sections []Section, name string) (Section, bool) {
	for _, section := range sections {
		if name != "" {
			if section.Title == name || section.Slug == name {
				return section, true
			}
			continue
		}
		for _, unit := range section.Units {
			if unit.HasCheck {
				return section, true
			}
		}
	}
	return Section{}, false
}

// parseHeading splits "## feat: title (slug)" into a Section; ok is false for any other heading.
func parseHeading(line string) (Section, bool) {
	text := strings.TrimSpace(strings.TrimLeft(line, "#"))
	sectionType, rest, found := strings.Cut(text, ": ")
	if !found || !isSectionType(sectionType) || !strings.HasSuffix(rest, ")") {
		return Section{}, false
	}
	open := strings.LastIndex(rest, " (")
	if open < 0 {
		return Section{}, false
	}
	return Section{
		Type:  sectionType,
		Title: rest[:open],
		Slug:  rest[open+2 : len(rest)-1],
	}, true
}

// isSectionType reports whether word is one of the Conventional Commits types a heading may use.
func isSectionType(word string) bool {
	for _, known := range sectionTypes {
		if word == known {
			return true
		}
	}
	return false
}

// isIndented reports whether line starts with a space or a tab.
func isIndented(line string) bool {
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
}

// hasCheck reports whether the unit's DONE line names one of checkNames.
func hasCheck(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		doneLine, isDone := strings.CutPrefix(strings.TrimSpace(line), "- DONE")
		if !isDone {
			continue
		}
		for _, name := range checkNames {
			if strings.Contains(doneLine, name) {
				return true
			}
		}
	}
	return false
}
