package git

import "strings"

// Change is one file the branch touched; Status is A (added), M (modified) or D (deleted).
type Change struct {
	Status string
	Path   string
}

// ChangedFiles lists the files dir's branch changed since it left base, a rename coming back as a D and an A.
// It returns an empty list when nothing changed, or an error when git fails.
func ChangedFiles(dir, base string) ([]Change, error) {
	output, err := runGit(dir, "diff", "--name-status", "--no-renames", base+"...HEAD")
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

// AddedLine is one line a branch added: the file it is in and its text without the leading "+".
type AddedLine struct {
	Path string
	Text string
}

// AddedLines returns every line dir's branch added since it left base, each with the file it is in.
// It returns an empty list when nothing was added, or an error when git fails.
func AddedLines(dir, base string) ([]AddedLine, error) {
	output, err := runGit(dir, "diff", "--unified=0", base+"...HEAD")
	if err != nil {
		return nil, err
	}
	var lines []AddedLine
	path := ""
	for line := range strings.SplitSeq(output, "\n") {
		if filePath, isFileHeader := strings.CutPrefix(line, "+++ b/"); isFileHeader {
			path = filePath
			continue
		}
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			lines = append(lines, AddedLine{Path: path, Text: line[1:]})
		}
	}
	return lines, nil
}
