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

// NewVersionLine returns "nightcap v0.1.7 is out (you have v0.1.6). Run `nightcap update`." when GitHub's latest release is not version.
// It returns "" for the same version, for a dev build, and when GitHub does not answer with a tag within 5 seconds.
func NewVersionLine(ctx context.Context, version string) string {
	if version == "dev" {
		return ""
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/repos/foreverfl/nightcap/releases/latest", nil)
	if err != nil {
		return ""
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return ""
	}
	defer response.Body.Close()
	var latest struct {
		TagName string `json:"tag_name"`
	}
	if response.StatusCode != http.StatusOK || json.NewDecoder(response.Body).Decode(&latest) != nil {
		return ""
	}
	if latest.TagName == "" || latest.TagName == version {
		return ""
	}
	return fmt.Sprintf("nightcap %s is out (you have %s). Run `nightcap update`.", latest.TagName, version)
}
