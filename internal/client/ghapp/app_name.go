package ghapp

import (
	"context"
	"encoding/json"
	"net/http"
)

// AppName asks GitHub which App the jwt belongs to and returns its slug, e.g. "nightcap-cli".
// It fails with GitHub's status and the start of its answer when the jwt is refused, e.g. "401 Unauthorized: …".
func AppName(ctx context.Context, jwt string) (string, error) {
	body, err := callGitHub(ctx, http.MethodGet, "/app", jwt)
	if err != nil {
		return "", err
	}
	var app struct {
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(body, &app); err != nil {
		return "", err
	}
	return app.Slug, nil
}
