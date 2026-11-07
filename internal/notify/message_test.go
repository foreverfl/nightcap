package notify

import (
	"strings"
	"testing"
	"time"
)

// oneNight is a night on blog-v2 from 23:05 to 03:40 with no usage lines.
func oneNight() Night {
	return Night{
		Repos: []string{"blog-v2"},
		Start: time.Date(2026, 11, 7, 23, 5, 0, 0, time.Local),
		End:   time.Date(2026, 11, 8, 3, 40, 0, 0, time.Local),
	}
}

func TestMessageWithUsageHasThreeLines(t *testing.T) {
	night := oneNight()
	night.Usage = []UsageLine{{"session", 9, 71}, {"weekly", 38, 52}}
	want := strings.Join([]string{
		"🌙 nightcap blog-v2 · 23:05 → 03:40",
		"  session 9% → 71%",
		"  weekly  38% → 52%",
	}, "\n")
	if got := Message(night); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

func TestMessageWithoutUsageIsTheFirstLineOnly(t *testing.T) {
	if got := Message(oneNight()); got != "🌙 nightcap blog-v2 · 23:05 → 03:40" {
		t.Fatalf("got %q", got)
	}
}
