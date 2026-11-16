package release

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// apiURL is GitHub's REST API; tests point it at a fake server.
var apiURL = "https://api.github.com"

// release is one entry of GitHub's release list, newest first.
type release struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Draft   bool   `json:"draft"`
}

// NewVersionNotice returns "nightcap v0.1.7 is out (you have v0.1.6). Run `nightcap update`." when GitHub's newest release is not version,
// followed by the feat, fix and refactor lines of every release above version.
// It returns nil for the same version, for a dev build, and when GitHub does not answer within 5 seconds.
func NewVersionNotice(ctx context.Context, version string) []string {
	if version == "dev" {
		return nil
	}
	releases, err := publishedReleases(ctx)
	if err != nil || len(releases) == 0 || releases[0].TagName == version {
		return nil
	}
	notice := []string{fmt.Sprintf("nightcap %s is out (you have %s). Run `nightcap update`.", releases[0].TagName, version)}
	return append(notice, changeLines(releases, version)...)
}

// publishedReleases returns nightcap's last 30 releases on GitHub, newest first, drafts left out.
func publishedReleases(ctx context.Context) ([]release, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/repos/foreverfl/nightcap/releases?per_page=30", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("releases: %s", response.Status)
	}
	var everyRelease []release
	if err := json.NewDecoder(response.Body).Decode(&everyRelease); err != nil {
		return nil, err
	}
	var published []release
	for _, entry := range everyRelease {
		if !entry.Draft {
			published = append(published, entry)
		}
	}
	return published, nil
}
