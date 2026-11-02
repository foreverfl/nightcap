package check

import (
	"context"
	"strings"
	"testing"
)

func TestRunTrueIsGreen(t *testing.T) {
	result, err := Run(context.Background(), t.TempDir(), "true")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Green {
		t.Fatal("the command `true` exits 0, so Green should be true, but it came back false")
	}
}

func TestRunFalseIsRed(t *testing.T) {
	result, err := Run(context.Background(), t.TempDir(), "false")
	if err != nil {
		t.Fatal(err)
	}
	if result.Green {
		t.Fatal("the command `false` exits 1, so Green should be false, but it came back true")
	}
}

func TestRunKeepsLast40Lines(t *testing.T) {
	result, err := Run(context.Background(), t.TempDir(), "seq 1 50")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(result.Tail, "\n")
	if len(lines) != 40 || lines[0] != "11" || lines[39] != "50" {
		t.Fatalf("seq 1 50 prints 50 lines, so Tail should be the 40 lines from 11 to 50; got %d lines, first %q, last %q", len(lines), lines[0], lines[len(lines)-1])
	}
}
