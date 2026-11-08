package notify

import (
	"fmt"
	"strings"
	"time"

	"nightcap/internal/plan"
)

// Summary is what one nightcap run did, as the morning message reports it.
type Summary struct {
	Repos      []string
	Start, End time.Time
	Usage      []UsageLine // empty when no commit ran or /usage could not be read
	PRs        []PR        // one per plan PR the night worked on, in plan order
	StopReason string      // why the run stopped early, such as "session 85% ≥ 80"; empty when it ran to the end
}

// PR is one PR of the plan at the end of the night.
type PR struct {
	Mark        plan.Mark
	Title       string // as the plan writes it, such as "feat: 검색 (search)"
	Done, Total int    // commits ended green out of the PR's commits
	URL         string // the PR link; empty when no PR was opened
	Failure     string // the first commit's result word when nothing ended green: red, blocked or stopped
}

// UsageLine is one /usage limit at the start and the end of the night.
type UsageLine struct {
	Name       string
	Start, End int
}

// Message builds the morning message: the header and usage lines, one line per PR, then why the run stopped early.
func Message(summary Summary) string {
	lines := []string{fmt.Sprintf("🌙 nightcap %s · %s → %s", strings.Join(summary.Repos, ", "), summary.Start.Format("15:04"), summary.End.Format("15:04"))}
	for _, limit := range summary.Usage {
		lines = append(lines, fmt.Sprintf("  %-7s %d%% → %d%%", limit.Name, limit.Start, limit.End))
	}
	if len(summary.PRs) > 0 {
		lines = append(lines, "")
	}
	for _, pr := range summary.PRs {
		lines = append(lines, prLine(pr))
	}
	if summary.StopReason != "" {
		lines = append(lines, "", "stopped: "+summary.StopReason)
	}
	return strings.Join(lines, "\n")
}

// prLine draws one PR as "<result> <mark> <title> · n/m · <state> · <link>"; a finished 🤖 PR reads "auto-merge", as nothing has merged yet.
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
