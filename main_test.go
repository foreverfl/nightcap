package main

import (
	"bytes"
	"os"
	"testing"
)

func TestDispatchRunPrintsPickedSection(t *testing.T) {
	t.Chdir(t.TempDir())
	planText := []byte("## feat: feed (feed)\n\n- [ ] add it\n  - DONE  go test passes\n")
	if err := os.WriteFile("plan.md", planText, 0o644); err != nil {
		t.Fatal(err)
	}
	yml := []byte("targets:\n  - plan: plan.md\n    repo: /tmp/repo\n    check: make check\n")
	if err := os.WriteFile("nightcap.yml", yml, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := dispatch([]string{"run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	want := "plan.md\n  branch: feat/feed\n  take  add it\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestDispatchUnknownCommandPrintsUsageAndExits2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := dispatch([]string{"foo"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr.String() != usage {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
