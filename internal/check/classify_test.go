package check

import (
	"testing"

	"nightcap/internal/git"
)

// protected and dependencies are the lists a target gets from nightcap.yml when it names none.
var protected = []string{"*_test.go", "*.hurl", "openapi/*", "Makefile"}
var dependencies = []string{"go.mod", "go.sum"}

func TestClassifyEditedTestFileIsRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "internal/feed/feed_test.go"}}, nil, nil, nil, protected, dependencies)
	if verdict != (Verdict{Red: true, Reason: "edited a contract file: internal/feed/feed_test.go"}) {
		t.Fatalf("editing an existing test file should be red, got %+v", verdict)
	}
}

func TestClassifyNewTestFileIsNotRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "A", Path: "internal/feed/feed_test.go"}}, nil, nil, nil, protected, dependencies)
	if verdict.Red {
		t.Fatalf("adding a new test file should not be red, got %+v", verdict)
	}
}

func TestClassifyEditedGoSumIsRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "go.sum"}}, nil, nil, nil, protected, dependencies)
	if verdict != (Verdict{Red: true, Reason: "changed a dependency file: go.sum"}) {
		t.Fatalf("changing go.sum should be red, got %+v", verdict)
	}
}

func TestClassifyAddedSkipLineIsRed(t *testing.T) {
	verdict := Classify(nil, []git.AddedLine{{Path: "feed_test.go", Text: `	t.Skip("later")`}}, nil, nil, protected, dependencies)
	if verdict != (Verdict{Red: true, Reason: `added a skip marker in feed_test.go: t.Skip("later")`}) {
		t.Fatalf("adding a t.Skip line should be red, got %+v", verdict)
	}
}

func TestClassifyAddedExitLineIsNotRed(t *testing.T) {
	verdict := Classify(nil, []git.AddedLine{{Path: "feed_test.go", Text: "	os.Exit(1)"}}, nil, nil, protected, dependencies)
	if verdict.Red {
		t.Fatalf("os.Exit( ends in xit( but is not a skip, got %+v", verdict)
	}
}

func TestClassifyAddedXitLineIsRed(t *testing.T) {
	verdict := Classify(nil, []git.AddedLine{{Path: "feed_test.go", Text: `  xit("loads the feed", () => {`}}, nil, nil, protected, dependencies)
	if !verdict.Red {
		t.Fatalf("a line starting with xit( should be red, got %+v", verdict)
	}
}

func TestClassifyFileOutsideWhereIsReadCode(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "main.go"}}, nil, []string{"internal/feed/feed.go"}, nil, protected, dependencies)
	if verdict != (Verdict{ReadCode: true, Reason: "outside WHERE: main.go"}) {
		t.Fatalf("a file outside WHERE should need its code read, got %+v", verdict)
	}
}

func TestClassifyFileMatchingReadCodeIsReadCode(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "internal/auth/auth.go"}}, nil, []string{"internal/auth/auth.go"}, []string{"auth/*"}, protected, dependencies)
	if verdict != (Verdict{ReadCode: true, Reason: "matches read_code: internal/auth/auth.go"}) {
		t.Fatalf("a file matching read_code should need its code read, got %+v", verdict)
	}
}

func TestClassifyFileInsideWhereIsGreen(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "internal/feed/feed.go"}}, nil, []string{"internal/feed/feed.go"}, nil, protected, dependencies)
	if verdict != (Verdict{}) {
		t.Fatalf("a file inside WHERE should give an empty Verdict, got %+v", verdict)
	}
}

func TestClassifyEditedTspIsRedWhenProtected(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "api.tsp"}}, nil, []string{"api.tsp"}, nil, []string{"*.tsp"}, nil)
	if verdict != (Verdict{Red: true, Reason: "edited a contract file: api.tsp"}) {
		t.Fatalf("editing api.tsp under protected *.tsp should be red, got %+v", verdict)
	}
}

func TestClassifyRegeneratedOpenapiIsNotRedWhenOnlyTspIsProtected(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "openapi.yaml"}}, nil, []string{"openapi.yaml"}, nil, []string{"*.tsp"}, nil)
	if verdict.Red {
		t.Fatalf("openapi.yaml is generated and should not be red, got %+v", verdict)
	}
}
