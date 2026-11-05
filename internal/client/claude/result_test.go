package claude

import (
	"encoding/json"
	"testing"
)

// oneResult is the smallest successful `claude -p --output-format json` output, as a map to tweak.
func oneResult() map[string]any {
	return map[string]any{"type": "result", "subtype": "success", "is_error": false, "result": "ok", "api_error_status": nil}
}

// parse encodes output and runs ParseResult on it.
func parse(t *testing.T, output map[string]any) Result {
	t.Helper()
	raw, _ := json.Marshal(output)
	result, err := ParseResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestParseResultDone(t *testing.T) {
	if got := parse(t, oneResult()).Outcome; got != Done {
		t.Fatalf("Outcome = %q, want done", got)
	}
}

func TestParseResultBlocked(t *testing.T) {
	output := oneResult()
	output["result"] = "blocked: needs a redis server"
	if got := parse(t, output).Outcome; got != Blocked {
		t.Fatalf("Outcome = %q, want blocked", got)
	}
}

func TestParseResultRateLimited(t *testing.T) {
	output := oneResult()
	output["is_error"] = true
	output["api_error_status"] = 429
	if got := parse(t, output).Outcome; got != RateLimited {
		t.Fatalf("Outcome = %q, want rate_limited", got)
	}
}

func TestParseResultRejectsPlainText(t *testing.T) {
	if _, err := ParseResult([]byte("ok")); err == nil {
		t.Fatal("expected error for non-JSON output")
	}
}

func TestParseResultKeepsTheSessionID(t *testing.T) {
	output := oneResult()
	output["session_id"] = "b2dcfa87-8868-4dda-9a16-71f70611bc18"
	if got := parse(t, output).SessionID; got != "b2dcfa87-8868-4dda-9a16-71f70611bc18" {
		t.Fatalf("SessionID = %q, want the session_id from the output", got)
	}
}
