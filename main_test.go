package main

import (
	"bytes"
	"testing"
)

func TestDispatchVersionPrintsTheBuiltInVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := dispatch([]string{"version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}
	if stdout.String() != "nightcap dev\n" {
		t.Fatalf("stdout = %q, want %q", stdout.String(), "nightcap dev\n")
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
