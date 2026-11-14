package verdict

import (
	"testing"

	"nightcap/internal/client/git"
)

// protected and dependencies are the lists a target gets from nightcap.yml when it names none.
var protected = []string{"*_test.go", "*.hurl", "openapi/*", "Makefile"}
var dependencies = []string{"go.mod", "go.sum"}

func TestClassifyEditedTestFileIsRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "internal/feed/feed_test.go"}}, nil, nil, protected, dependencies)
	if verdict != (Result{Red: true, Reason: "edited a contract file: internal/feed/feed_test.go"}) {
		t.Fatalf("editing an existing test file should be red, got %+v", verdict)
	}
}

func TestClassifyNewTestFileIsNotRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "A", Path: "internal/feed/feed_test.go"}}, nil, nil, protected, dependencies)
	if verdict.Red {
		t.Fatalf("adding a new test file should not be red, got %+v", verdict)
	}
}

func TestClassifyEditedGoSumIsRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "go.sum"}}, nil, nil, protected, dependencies)
	if verdict != (Result{Red: true, Reason: "changed a dependency file: go.sum"}) {
		t.Fatalf("changing go.sum should be red, got %+v", verdict)
	}
}

func TestClassifyAddedSkipLineIsRed(t *testing.T) {
	verdict := Classify(nil, []git.AddedLine{{Path: "feed_test.go", Text: `	t.Skip("later")`}}, nil, protected, dependencies)
	if verdict != (Result{Red: true, Reason: `added a skip marker in feed_test.go: t.Skip("later")`}) {
		t.Fatalf("adding a t.Skip line should be red, got %+v", verdict)
	}
}

func TestClassifyAddedExitLineIsNotRed(t *testing.T) {
	verdict := Classify(nil, []git.AddedLine{{Path: "feed_test.go", Text: "	os.Exit(1)"}}, nil, protected, dependencies)
	if verdict.Red {
		t.Fatalf("os.Exit( ends in xit( but is not a skip, got %+v", verdict)
	}
}

func TestClassifyAddedXitLineIsRed(t *testing.T) {
	verdict := Classify(nil, []git.AddedLine{{Path: "feed_test.go", Text: `  xit("loads the feed", () => {`}}, nil, protected, dependencies)
	if !verdict.Red {
		t.Fatalf("a line starting with xit( should be red, got %+v", verdict)
	}
}

func TestClassifyFileOutsideWhereIsReadCode(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "main.go"}}, nil, []string{"internal/feed/feed.go"}, protected, dependencies)
	if verdict != (Result{ReadCode: true, Reason: "outside WHERE: main.go"}) {
		t.Fatalf("a file outside WHERE should need its code read, got %+v", verdict)
	}
}

func TestClassifyFileInsideWhereIsGreen(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "internal/feed/feed.go"}}, nil, []string{"internal/feed/feed.go"}, protected, dependencies)
	if verdict != (Result{}) {
		t.Fatalf("a file inside WHERE should give an empty Result, got %+v", verdict)
	}
}

func TestClassifyEditedTspIsRedWhenProtected(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "api.tsp"}}, nil, []string{"api.tsp"}, []string{"*.tsp"}, nil)
	if verdict != (Result{Red: true, Reason: "edited a contract file: api.tsp"}) {
		t.Fatalf("editing api.tsp under protected *.tsp should be red, got %+v", verdict)
	}
}

func TestClassifyRegeneratedOpenapiIsNotRedWhenOnlyTspIsProtected(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "openapi.yaml"}}, nil, []string{"openapi.yaml"}, []string{"*.tsp"}, nil)
	if verdict.Red {
		t.Fatalf("openapi.yaml is generated and should not be red, got %+v", verdict)
	}
}

func TestClassifySourceFileUnderStarFolderIsNotRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "services/rust/src/github/handler.rs"}}, nil, []string{"services/rust/src/github/handler.rs"}, nil, []string{"services/*/Cargo.toml"})
	if verdict != (Result{}) {
		t.Fatalf("handler.rs is not services/*/Cargo.toml and should give an empty Result, got %+v", verdict)
	}
}

func TestClassifyCargoTomlUnderStarFolderIsRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "services/rust/Cargo.toml"}}, nil, nil, nil, []string{"services/*/Cargo.toml"})
	if verdict != (Result{Red: true, Reason: "changed a dependency file: services/rust/Cargo.toml"}) {
		t.Fatalf("services/rust/Cargo.toml should be red, got %+v", verdict)
	}
}

func TestClassifyTestFileUnderDoubleStarIsRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "services/rust/src/github/handler_test.go"}}, nil, nil, []string{"services/**/*_test.go"}, nil)
	if verdict != (Result{Red: true, Reason: "edited a contract file: services/rust/src/github/handler_test.go"}) {
		t.Fatalf("a test file deep under services/** should be red, got %+v", verdict)
	}
}

func TestClassifySpecUnderTrailingDoubleStarIsRed(t *testing.T) {
	verdict := Classify([]git.Change{{Status: "M", Path: "doc-source/openapi/specs/a/b.tsp"}}, nil, nil, []string{"doc-source/openapi/specs/**"}, nil)
	if verdict != (Result{Red: true, Reason: "edited a contract file: doc-source/openapi/specs/a/b.tsp"}) {
		t.Fatalf("a file deep under specs/** should be red, got %+v", verdict)
	}
}
