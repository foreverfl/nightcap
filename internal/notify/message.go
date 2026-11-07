package notify

import (
	"fmt"
	"strings"
	"time"

	"nightcap/internal/plan"
)

// Night is what one nightcap run did, as the morning message reports it.
type Night struct {
	Repos      []string
	Start, End time.Time
	Usage      []UsageLine // empty when no unit ran or /usage could not be read
	PRs        []PR        // one per plan heading the night worked on, in plan order
}

// PR is one plan heading, which is one PR, at the end of the night.
type PR struct {
	Mark        plan.Mark
	Title       string // as the plan writes it, such as "feat: 검색 (search)"
	Done, Total int    // units ended green out of the heading's units
	URL         string // the PR link; empty when no PR was opened
	Failure     string // the first unit's result word when nothing ended green: red, blocked or stopped
}

// UsageLine is one /usage limit at the start and the end of the night.
type UsageLine struct {
	Name       string
	Start, End int
}

// Message builds the end-of-night message: the repos and times, one indented line per usage limit,
// then after a blank line one line per PR.
func Message(night Night) string {
	lines := []string{fmt.Sprintf("🌙 nightcap %s · %s → %s", strings.Join(night.Repos, ", "), night.Start.Format("15:04"), night.End.Format("15:04"))}
	for _, limit := range night.Usage {
		lines = append(lines, fmt.Sprintf("  %-7s %d%% → %d%%", limit.Name, limit.Start, limit.End))
	}
	if len(night.PRs) > 0 {
		lines = append(lines, "")
	}
	for _, pr := range night.PRs {
		lines = append(lines, prLine(pr))
	}
	return strings.Join(lines, "\n")
}

// prLine draws one PR as "<result> <mark> <title> · n/m · <state> · <link>".
// The message goes out before GitHub merges anything, so a finished 🤖 heading reads "auto-merge", never "merged".
func prLine(pr PR) string {
	result, state, link := "✅", "open", pr.URL
	switch {
	case pr.Done == 0:
		result, state, link = "🔴", pr.Failure, "no PR"
	case pr.Done < pr.Total:
		result = "🟡"
	case pr.Mark == plan.MarkAuto:
		state = "auto-merge"
	}
	return fmt.Sprintf("%s %s %s · %d/%d · %s · %s", result, pr.Mark.Emoji(), pr.Title, pr.Done, pr.Total, state, link)
}
