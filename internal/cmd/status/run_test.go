package status

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatusPrintsTheFrontmatterRepo(t *testing.T) {
	repo := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		if output, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	planPath := filepath.Join(t.TempDir(), "commit-bot.md")
	if err := os.WriteFile(planPath, []byte("---\nrepo: "+repo+"\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	// each worktree line is git worktree list's own "<path>  <commit> [main]"
	if err := Run([]string{planPath}, &stdout); err != nil || !strings.HasPrefix(stdout.String(), repo+"\n") || !strings.Contains(stdout.String(), "[main]") {
		t.Fatalf("want the repo line, then its main worktree; got %v %q", err, stdout.String())
	}
}

func TestStatusWithoutAPlanFileFails(t *testing.T) {
	var stdout bytes.Buffer
	if err := Run(nil, &stdout); err == nil || err.Error() != "status: name a plan file" {
		t.Fatalf("want status: name a plan file, got %v", err)
	}
}
