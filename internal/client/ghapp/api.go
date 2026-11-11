package ghapp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// apiURL is GitHub's REST API; tests point it at a fake server.
var apiURL = "https://api.github.com"

// callGitHub sends method to apiURL+path with the jwt as a Bearer token and returns the answer's body.
// It fails with GitHub's status and the start of its answer on anything but a 2xx, e.g. "401 Unauthorized: …".
func callGitHub(ctx context.Context, method, path, jwt string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, method, apiURL+path, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+jwt)
	client := http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%s: %.200s", response.Status, body)
	}
	return body, nil
}
