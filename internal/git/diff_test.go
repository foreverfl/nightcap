package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// commitFile writes content to name inside dir and commits it on dir's branch.
func commitFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", name},
		{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "add " + name},
	} {
		if output, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
}

func TestChangedFilesListsOneAddedFile(t *testing.T) {
	dir, err := AddWorktree(oneRepo(t), "feat/feed", "feed")
	if err != nil {
		t.Fatal(err)
	}
	commitFile(t, dir, "feed.go", "package feed\n")
	changes, err := ChangedFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0] != (Change{Status: "A", Path: "feed.go"}) {
		t.Fatalf("one committed new file should come back as [{A feed.go}], got %v", changes)
	}
}

func TestAddedLinesReturnsTheOneAddedLine(t *testing.T) {
	dir, err := AddWorktree(oneRepo(t), "feat/feed", "feed")
	if err != nil {
		t.Fatal(err)
	}
	commitFile(t, dir, "feed.go", "package feed\n")
	lines, err := AddedLines(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0] != "package feed" {
		t.Fatalf("a commit adding one line should come back as [package feed], got %q", lines)
	}
}
