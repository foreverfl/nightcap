package notify

import (
	"strings"
	"testing"
	"time"

	"nightcap/internal/plan"
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

func TestMessageDrawsOneLinePerPR(t *testing.T) {
	night := oneNight()
	night.PRs = []PR{
		{Mark: plan.MarkAuto, Title: "test: 피드 hurl 파일 (feed-hurl)", Done: 3, Total: 3, URL: "https://github.com/foreverfl/blog-v2/pull/41"},
		{Mark: plan.MarkReadPR, Title: "feat: 검색 (search)", Done: 0, Total: 2, Failure: "blocked"},
		{Mark: plan.MarkReadPR, Title: "fix: 빈 피드 (empty-feed)", Done: 1, Total: 2, URL: "https://github.com/foreverfl/blog-v2/pull/43"},
	}
	want := strings.Join([]string{
		"🌙 nightcap blog-v2 · 23:05 → 03:40",
		"",
		"✅ 🤖 test: 피드 hurl 파일 (feed-hurl) · 3/3 · auto-merge · https://github.com/foreverfl/blog-v2/pull/41",
		"🔴 📄 feat: 검색 (search) · 0/2 · blocked · no PR",
		"🟡 📄 fix: 빈 피드 (empty-feed) · 1/2 · open · https://github.com/foreverfl/blog-v2/pull/43",
	}, "\n")
	if got := Message(night); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}
