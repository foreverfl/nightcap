package usage

import (
	"bytes"
	"encoding/json"
	"time"
)

// Limit is one /usage limit: the percent used and when it resets.
type Limit struct {
	Percent  int
	ResetsAt time.Time
}

// Usage is the session and weekly_all limits from /usage; Known is false when the output had no rate_limits.
type Usage struct {
	Session   Limit
	WeeklyAll Limit
	Known     bool
}

// Read takes `claude -p /usage --output-format stream-json --verbose` output, finds the line with usage_report,
// and returns its session and weekly_all limits; weekly_scoped is skipped because night runs never use that model.
func Read(output []byte) (Usage, error) {
	for line := range bytes.SplitSeq(output, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		// matches {"usage_report":{"rate_limits":{"limits":[{"kind","percent","resets_at"}]}}}; other keys are dropped
		var event struct {
			UsageReport *struct {
				RateLimits *struct {
					Limits []struct {
						Kind     string    `json:"kind"`
						Percent  int       `json:"percent"`
						ResetsAt time.Time `json:"resets_at"`
					} `json:"limits"`
				} `json:"rate_limits"`
			} `json:"usage_report"`
		}
		if err := json.Unmarshal(line, &event); err != nil {
			return Usage{}, err
		}
		if event.UsageReport == nil {
			continue
		}
		if event.UsageReport.RateLimits == nil {
			return Usage{}, nil
		}
		usage := Usage{Known: true}
		for _, limit := range event.UsageReport.RateLimits.Limits {
			switch limit.Kind {
			case "session":
				usage.Session = Limit{Percent: limit.Percent, ResetsAt: limit.ResetsAt}
			case "weekly_all":
				usage.WeeklyAll = Limit{Percent: limit.Percent, ResetsAt: limit.ResetsAt}
			}
		}
		return usage, nil
	}
	return Usage{}, nil
}
