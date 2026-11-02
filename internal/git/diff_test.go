package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestChangedFilesListsOneAddedFile(t *testing.T) {
	dir, err := AddWorktree(oneRepo(t), "feat/feed", "feed")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "feed.go"), []byte("package feed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "feed.go"},
		{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "add feed"},
	} {
		if output, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	changes, err := ChangedFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0] != (Change{Status: "A", Path: "feed.go"}) {
		t.Fatalf("one committed new file should come back as [{A feed.go}], got %v", changes)
	}
}
