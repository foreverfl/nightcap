package verdict

import (
	"path"
	"strings"

	"nightcap/internal/git"
)

// Result is what the branch's changes say about a commit whose test passed; Red means the commit is not accepted,
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

// Classify returns a red Result when a change touches a dependency file, edits or deletes a protected file, or adds a skip marker,
// a ReadCode one when it adds a migration or sits outside wherePaths, and an empty Result otherwise.
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

// isInsideWhere reports whether filePath is a WHERE path, sits under one as a folder, or shares a file name with one that names no folder.
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

// matchesAny reports whether filePath matches one of patterns; a pattern without "/" matches the file name, one with "/" the whole path folder by folder.
func matchesAny(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(pattern, "/") {
			if isWholePathMatch(strings.Split(pattern, "/"), strings.Split(filePath, "/")) {
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

// isWholePathMatch reports whether pathFolders match patternFolders one folder at a time, where "**" stands for any number of folders.
func isWholePathMatch(patternFolders, pathFolders []string) bool {
	if len(patternFolders) == 0 {
		return len(pathFolders) == 0
	}
	// "**" first: drop it and try the rest against the path with 0, 1, 2… folders skipped.
	if patternFolders[0] == "**" {
		for skipped := 0; skipped <= len(pathFolders); skipped++ {
			if isWholePathMatch(patternFolders[1:], pathFolders[skipped:]) {
				return true
			}
		}
		return false
	}
	if len(pathFolders) == 0 {
		return false
	}
	// Anything else: compare the first folders, then drop one from each and compare the rest.
	matched, _ := path.Match(patternFolders[0], pathFolders[0])
	return matched && isWholePathMatch(patternFolders[1:], pathFolders[1:])
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
