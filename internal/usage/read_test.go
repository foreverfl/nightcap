package usage

import (
	"os"
	"testing"
	"time"
)

func TestReadTakesSessionAndWeeklyAll(t *testing.T) {
	output, err := os.ReadFile("testdata/usage.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Read(output)
	if err != nil {
		t.Fatal(err)
	}
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	sessionReset := time.Date(2026, 11, 5, 4, 40, 0, 188163000, tokyo)
	weeklyReset := time.Date(2026, 11, 9, 8, 0, 0, 188185000, tokyo)
	if !got.Known || got.Session.Percent != 15 || !got.Session.ResetsAt.Equal(sessionReset) ||
		got.WeeklyAll.Percent != 35 || !got.WeeklyAll.ResetsAt.Equal(weeklyReset) {
		t.Fatalf("got %+v, want session 15%% until %v and weekly_all 35%% until %v", got, sessionReset, weeklyReset)
	}
}

func TestReadNullRateLimitsIsUnknown(t *testing.T) {
	got, err := Read([]byte(`{"type":"assistant","usage_report":{"rate_limits":null}}`))
	if err != nil || got.Known {
		t.Fatalf("got %+v, %v; want Known false and no error", got, err)
	}
}

func TestReadBrokenJSONFails(t *testing.T) {
	if _, err := Read([]byte(`{"type":`)); err == nil {
		t.Fatal("want an error for broken JSON")
	}
}
