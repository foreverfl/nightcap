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

var dependencyFiles = []string{
	"go.mod", "go.sum",
	"Cargo.toml", "Cargo.lock",
	"package.json", "package-lock.json", "pnpm-lock.yaml", "yarn.lock",
}

// Classify returns a red Verdict when a change touches a dependency file in any way,
// or modifies or deletes an existing contract file; adding a new contract file is allowed.
// It returns an empty Verdict when no rule matches.
func Classify(changes []git.Change) Verdict {
	for _, change := range changes {
		if matchesAny(change.Path, dependencyFiles) {
			return Verdict{Red: true, Reason: "changed a dependency file: " + change.Path}
		}
		if change.Status != "A" && matchesAny(change.Path, contractFiles) {
			return Verdict{Red: true, Reason: "edited a contract file: " + change.Path}
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
