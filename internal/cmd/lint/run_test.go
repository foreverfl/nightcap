package lint

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// oneLintPlan is the smallest plan lint passes: one typed PR with one commit that has a test.
func oneLintPlan() string {
	return "## 📄 feat: a (b)\n" +
		"- [ ] read the docs\n" +
		"  - DONE  go test ./... passes\n"
}

// runLint writes plan to a temp file, runs lint on it, and returns its error and stdout.
func runLint(t *testing.T, plan string) (error, string) {
	t.Helper()
	planPath := filepath.Join(t.TempDir(), "commit-bot.md")
	if err := os.WriteFile(planPath, []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	err := Run([]string{planPath}, &stdout)
	return err, stdout.String()
}

func TestLintPlanWithoutTypedPRFails(t *testing.T) {
	err, stdout := runLint(t, strings.Replace(oneLintPlan(), "## 📄 feat: a (b)", "## cmd", 1))
	if err == nil || stdout != "no PR with type: slug\n" {
		t.Fatalf("want an error and the no-PR line, got %v %q", err, stdout)
	}
}

func TestLintCommitWithoutTestFails(t *testing.T) {
	err, stdout := runLint(t, strings.Replace(oneLintPlan(), "go test ./... passes", "테스트가 없음", 1))
	want := "PR: feat: a (b)\n  no test: read the docs\n"
	if err == nil || stdout != want {
		t.Fatalf("want an error and %q, got %v %q", want, err, stdout)
	}
}

func TestLintPrintsAPersonPRAsPerson(t *testing.T) {
	err, stdout := runLint(t, strings.Replace(oneLintPlan(), "## 📄 feat:", "## 👤 feat:", 1))
	if err != nil || stdout != "person: feat: a (b)\n" {
		t.Fatalf("want no error and %q, got %v %q", "person: feat: a (b)\n", err, stdout)
	}
}

func TestLintPRWithoutAMarkFails(t *testing.T) {
	err, stdout := runLint(t, strings.Replace(oneLintPlan(), "📄 ", "", 1))
	if err == nil || stdout != "no mark: feat: a (b)\n" {
		t.Fatalf("want an error and %q, got %v %q", "no mark: feat: a (b)\n", err, stdout)
	}
}

func TestLintWherePathWithASpaceFails(t *testing.T) {
	err, stdout := runLint(t, strings.Replace(oneLintPlan(), "  - DONE", "  - WHERE (blog-v2) a b/\n  - DONE", 1))
	want := "PR: feat: a (b)\n  WHERE path not readable: a b/\n"
	if err == nil || stdout != want {
		t.Fatalf("want an error and %q, got %v %q", want, err, stdout)
	}
}
