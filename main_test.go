package main

import (
	"bytes"
	"testing"
)

func TestDispatchRunPrintsRepoPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := dispatch([]string{"run", "/tmp/x"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	if stdout.String() != "repo: /tmp/x\n" {
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
