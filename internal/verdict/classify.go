package verdict

import (
	"path"
	"strings"

	"nightcap/internal/git"
)

// Verdict is what the branch's changes say about a green unit; Red means the commit is not accepted,
// ReadCode means it is accepted but the PR needs its code read.
type Result struct {
	Red      bool
	ReadCode bool
	Reason   string
}

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

// A newly added file whose path contains this word is treated as a database migration.
const migrationPathWord = "migration"

// Classify returns a red Verdict when a change touches one of dependencies, modifies or deletes an existing
// file matching protected (adding a new one is allowed), or an added line carries a skip marker; past those, a ReadCode
// one when a change adds a migration or sits outside wherePaths. Otherwise an empty Verdict.
func Classify(changes []git.Change, addedLines []git.AddedLine, wherePaths, protected, dependencies []string) Result {
	for _, change := range changes {
		if matchesAny(change.Path, dependencies) {
			return Result{Red: true, Reason: "changed a dependency file: " + change.Path}
		}
		if change.Status != "A" && matchesAny(change.Path, protected) {
			return Result{Red: true, Reason: "edited a contract file: " + change.Path}
		}
	}
	for _, line := range addedLines {
		if hasSkipMarker(line.Text) {
			return Result{Red: true, Reason: "added a skip marker in " + line.Path + ": " + strings.TrimSpace(line.Text)}
		}
	}
	for _, change := range changes {
		if change.Status == "A" && strings.Contains(change.Path, migrationPathWord) {
			return Result{ReadCode: true, Reason: "added a migration: " + change.Path}
		}
		if !isInsideWhere(change.Path, wherePaths) {
			return Result{ReadCode: true, Reason: "outside WHERE: " + change.Path}
		}
	}
	return Result{}
}

// isInsideWhere reports whether filePath is one of wherePaths, sits under one of them as a folder,
// or has the same file name as an entry that names no folder.
func isInsideWhere(filePath string, wherePaths []string) bool {
	for _, where := range wherePaths {
		folder := strings.TrimSuffix(where, "/")
		hasNoFolder := !strings.Contains(where, "/")
		if filePath == where || strings.HasPrefix(filePath, folder+"/") || (hasNoFolder && path.Base(filePath) == where) {
			return true
		}
	}
	return false
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
