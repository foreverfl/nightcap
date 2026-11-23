package plan

import (
	"bufio"
	"slices"
	"strings"
)

// prTypes are the Conventional Commits types a PR may use.
var prTypes = []string{"feat", "fix", "test", "docs", "refactor", "chore"}

// Areas are the words a unit's "[repo/area]" tag may use after the slash.
var Areas = []string{"front", "back", "cli", "sql", "infra", "config", "docs", "test", "none"}

// marks maps each reading mark written before a PR's type to its Mark.
var marks = map[string]Mark{"🤖": MarkAuto, "📄": MarkReadPR, "👤": MarkPerson}

var testNames = []string{"hurl", "go test", "make check", "cargo test", "npm test"}

// Parse returns every "## type: title (slug)" PR in a plan file with its open commits, in file order.
// A "##" line without a known type or a slug is skipped together with the commits under it.
func Parse(text string) []PR {
	var planPRs []PR
	var current *Commit
	underTypedPR := false
	inFence := false

	flush := func() {
		if current == nil {
			return
		}
		current.Body = strings.TrimRight(current.Body, "\n")
		current.HasTest = hasTest(current.Body)
		last := &planPRs[len(planPRs)-1]
		last.Commits = append(last.Commits, *current)
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
			planPR, ok := parsePRLine(line)
			underTypedPR = ok
			if ok {
				planPRs = append(planPRs, planPR)
			}
			continue
		}
		if !underTypedPR {
			continue
		}
		if taggedTitle, ok := strings.CutPrefix(line, "- [ ] "); ok {
			repo, area, title := splitTag(taggedTitle)
			current = &Commit{Repo: repo, Area: area, Title: title, Body: "- [ ] " + title + "\n"}
			continue
		}
		last := &planPRs[len(planPRs)-1]
		isPlainLine := line != "" && !isIndented(line) && !strings.HasPrefix(line, "- ")
		if last.Summary == "" && len(last.Commits) == 0 && isPlainLine {
			last.Summary = line
		}
	}
	flush()
	return planPRs
}

// Pick returns the first PR that has a commit a night run can take, never one marked 👤 or not marked at all; ok is false when none matches.
func Pick(planPRs []PR) (PR, bool) {
	for _, planPR := range planPRs {
		if planPR.IsPerson() || planPR.Mark == MarkNone {
			continue
		}
		for _, commit := range planPR.Commits {
			if commit.HasTest {
				return planPR, true
			}
		}
	}
	return PR{}, false
}

// parsePRLine splits "## 📄 feat: title (slug)" into a PR; ok is false without a known type or a slug.
func parsePRLine(line string) (PR, bool) {
	text := strings.TrimSpace(strings.TrimLeft(line, "#"))
	mark := MarkNone
	if symbol, rest, found := strings.Cut(text, " "); found && marks[symbol] != MarkNone {
		mark, text = marks[symbol], rest
	}
	prType, rest, found := strings.Cut(text, ": ")
	if !found || !slices.Contains(prTypes, prType) || !strings.HasSuffix(rest, ")") {
		return PR{}, false
	}
	open := strings.LastIndex(rest, " (")
	if open < 0 {
		return PR{}, false
	}
	return PR{
		Mark:  mark,
		Type:  prType,
		Title: rest[:open],
		Slug:  rest[open+2 : len(rest)-1],
	}, true
}

// splitTag splits "[blog-v2/back] add GET /github/kinds" into "blog-v2", "back" and "add GET /github/kinds".
// A title that does not start with "[repo/area] " comes back whole, with repo and area empty.
func splitTag(title string) (repo, area, rest string) {
	if !strings.HasPrefix(title, "[") {
		return "", "", title
	}
	tag, rest, isClosed := strings.Cut(title[1:], "] ")
	if !isClosed || strings.Count(tag, "/") != 1 {
		return "", "", title
	}
	repo, area, _ = strings.Cut(tag, "/")
	return repo, area, rest
}

// isIndented reports whether line starts with a space or a tab.
func isIndented(line string) bool {
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
}

// hasTest reports whether the commit's DONE line names one of testNames.
func hasTest(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		doneLine, isDone := strings.CutPrefix(strings.TrimSpace(line), "- DONE")
		if !isDone {
			continue
		}
		for _, name := range testNames {
			if strings.Contains(doneLine, name) {
				return true
			}
		}
	}
	return false
}
