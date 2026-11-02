package git

import "strings"

// Change is one file the branch touched; Status is A (added), M (modified) or D (deleted).
type Change struct {
	Status string
	Path   string
}

// ChangedFiles lists the files dir's branch changed since it left main, a rename coming back as a D and an A.
// It returns an empty list when nothing changed, or an error when git fails.
func ChangedFiles(dir string) ([]Change, error) {
	// TODO: repos that branch from staging (blog-v2) need the branch the worktree left from, not main.
	output, err := runGit(dir, "diff", "--name-status", "--no-renames", "main...HEAD")
	if err != nil {
		return nil, err
	}
	var changes []Change
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		status, path, found := strings.Cut(line, "\t")
		if !found {
			continue
		}
		changes = append(changes, Change{Status: status, Path: path})
	}
	return changes, nil
}
