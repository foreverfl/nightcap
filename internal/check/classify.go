package check

import (
	"path"
	"strings"

	"nightcap/internal/git"
)

// Verdict is what the branch's changes say about a green unit; Red means the commit is not accepted.
type Verdict struct {
	Red    bool
	Reason string
}

var contractFiles = []string{"*_test.go", "*.hurl", "openapi/*", "Makefile"}

var skipMarkers = []string{
	// Go testing
	"t.Skip",

	// Rust testing
	"#[ignore]",

	// JS / TS test skipping
	".skip(", ".skipIf(", ".only(",

	// Lint suppression
	"nolint", "eslint-disable",

	// Coverage exclusion
	"istanbul ignore", "c8 ignore", "pragma: no cover",

	// Generic options
	"skip: true",
}

// Bare function names like xit( also end other names (os.Exit(, benefit(), so they only count at the start of a line.
var skipCallsAtLineStart = []string{
	// JS / TS test skipping
	"xit(", "xtest(", "xdescribe(",

	// JS / TS test focusing
	"fit(", "fdescribe(",
}

var dependencyFiles = []string{
	"go.mod", "go.sum",
	"Cargo.toml", "Cargo.lock",
	"package.json", "package-lock.json", "pnpm-lock.yaml", "yarn.lock",
}

// Classify returns a red Verdict when a change touches a dependency file, modifies or deletes an existing
// contract file (adding a new one is allowed), or an added line carries a skip marker.
// It returns an empty Verdict when no rule matches.
func Classify(changes []git.Change, addedLines []string) Verdict {
	for _, change := range changes {
		if matchesAny(change.Path, dependencyFiles) {
			return Verdict{Red: true, Reason: "changed a dependency file: " + change.Path}
		}
		if change.Status != "A" && matchesAny(change.Path, contractFiles) {
			return Verdict{Red: true, Reason: "edited a contract file: " + change.Path}
		}
	}
	for _, line := range addedLines {
		if hasSkipMarker(line) {
			return Verdict{Red: true, Reason: "added a skip marker: " + strings.TrimSpace(line)}
		}
	}
	return Verdict{}
}

// matchesAny reports whether filePath matches one of patterns.
// A pattern without "/" is matched against the file name; "dir/*" matches any file under a dir folder.
func matchesAny(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		folder, _, hasFolder := strings.Cut(pattern, "/")
		if hasFolder {
			if strings.HasPrefix(filePath, folder+"/") || strings.Contains(filePath, "/"+folder+"/") {
				return true
			}
			continue
		}
		if matched, _ := path.Match(pattern, path.Base(filePath)); matched {
			return true
		}
	}
	return false
}

// hasSkipMarker reports whether line contains one of skipMarkers or starts with one of skipCallsAtLineStart.
func hasSkipMarker(line string) bool {
	for _, marker := range skipMarkers {
		if strings.Contains(line, marker) {
			return true
		}
	}
	trimmed := strings.TrimSpace(line)
	for _, call := range skipCallsAtLineStart {
		if strings.HasPrefix(trimmed, call) {
			return true
		}
	}
	return false
}
