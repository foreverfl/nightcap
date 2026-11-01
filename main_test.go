package main

import (
	"bytes"
	"os"
	"testing"
)

func TestDispatchRunPrintsTargets(t *testing.T) {
	t.Chdir(t.TempDir())
	yml := []byte("targets:\n  - plan: docs/plan.md\n    repo: /tmp/repo\n    check: make check\n")
	if err := os.WriteFile("nightcap.yml", yml, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := dispatch([]string{"run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	want := "max_turns: 30  max_rounds: 3  work_start: 09:00\ntarget: docs/plan.md -> /tmp/repo (make check)\n"
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
