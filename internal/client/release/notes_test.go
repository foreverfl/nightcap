package release

import (
	"slices"
	"strings"
	"testing"
)

func TestChangeLinesKeepsFeatAndFixAboveMyVersionWithoutHashes(t *testing.T) {
	releases := []release{
		{TagName: "v0.1.6", Body: strings.Join([]string{
			"## 🚀 Features",
			"",
			"- feat: let nightcap.yml set the author (0b5ac8ef45b8c45ec7a0a776359981020f6d3d87)",
			"",
			"## 📝 Documentation",
			"",
			"- docs: list where nightcap keeps its files (01ec317833b2751c71256dc87b5e8ca3089f752b)",
			"- Merge pull request #38 from foreverfl/feat/run-log (50a53601cac6f0076d020a61dcc9c743abe12d58)",
		}, "\n")},
		{TagName: "v0.1.5", Body: "- fix: open the PR with an English title (c4964ca78fbc7bd270154fe54d4ef56bc0adb65e)"},
		{TagName: "v0.1.4", Body: "- fix: an old fix (bebb89874b25ae7415ea6db8cb3d1f38f0db5e08)"},
	}
	want := []string{
		"- feat: let nightcap.yml set the author",
		"- fix: open the PR with an English title",
	}
	if got := changeLines(releases, "v0.1.4"); !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}
