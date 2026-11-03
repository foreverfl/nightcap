package check

import (
	"testing"

	"nightcap/internal/git"
)

func TestClassifyEditedTestFileIsRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "internal/feed/feed_test.go"}}, nil)
	if verdict != (Verdict{Red: true, Reason: "edited a contract file: internal/feed/feed_test.go"}) {
		t.Fatalf("editing an existing test file should be red, got %+v", verdict)
	}
}

func TestClassifyNewTestFileIsNotRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "A", Path: "internal/feed/feed_test.go"}}, nil)
	if verdict.Red {
		t.Fatalf("adding a new test file should not be red, got %+v", verdict)
	}
}

func TestClassifyEditedGoSumIsRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "go.sum"}}, nil)
	if verdict != (Verdict{Red: true, Reason: "changed a dependency file: go.sum"}) {
		t.Fatalf("changing go.sum should be red, got %+v", verdict)
	}
}

func TestClassifyAddedSkipLineIsRed(t *testing.T) {
	verdict := Classify(nil, []string{`	t.Skip("later")`})
	if verdict != (Verdict{Red: true, Reason: `added a skip marker: t.Skip("later")`}) {
		t.Fatalf("adding a t.Skip line should be red, got %+v", verdict)
	}
}

func TestClassifyAddedExitLineIsNotRed(t *testing.T) {
	verdict := Classify(nil, []string{"	os.Exit(1)"})
	if verdict.Red {
		t.Fatalf("os.Exit( ends in xit( but is not a skip, got %+v", verdict)
	}
}

func TestClassifyAddedXitLineIsRed(t *testing.T) {
	verdict := Classify(nil, []string{`  xit("loads the feed", () => {`})
	if !verdict.Red {
		t.Fatalf("a line starting with xit( should be red, got %+v", verdict)
	}
}
