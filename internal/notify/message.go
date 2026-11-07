package notify

import (
	"fmt"
	"strings"
	"time"
)

// Night is what one nightcap run did, as the morning message reports it.
type Night struct {
	Repos      []string
	Start, End time.Time
	Usage      []UsageLine // empty when no unit ran or /usage could not be read
}

// UsageLine is one /usage limit at the start and the end of the night.
type UsageLine struct {
	Name       string
	Start, End int
}

// Message builds the end-of-night message: the repos and times, then one indented line per usage limit.
func Message(night Night) string {
	lines := []string{fmt.Sprintf("🌙 nightcap %s · %s → %s", strings.Join(night.Repos, ", "), night.Start.Format("15:04"), night.End.Format("15:04"))}
	for _, limit := range night.Usage {
		lines = append(lines, fmt.Sprintf("  %-7s %d%% → %d%%", limit.Name, limit.Start, limit.End))
	}
	return strings.Join(lines, "\n")
}
