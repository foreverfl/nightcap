package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Send posts text as one message to a Discord or Slack webhook url, and does nothing when url is empty.
// It returns an error when the request fails or the webhook answers with a non-2xx status.
func Send(url, text string) error {
	if url == "" {
		return nil
	}
	body, err := json.Marshal(map[string]string{bodyKey(url): text})
	if err != nil {
		return err
	}
	client := http.Client{Timeout: 10 * time.Second}
	response, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("webhook answered %s", response.Status)
	}
	return nil
}

// bodyKey returns the JSON key the webhook reads the message from: "text" for Slack, "content" for Discord.
func bodyKey(url string) string {
	if strings.HasPrefix(url, "https://hooks.slack.com/") {
		return "text"
	}
	return "content"
}
