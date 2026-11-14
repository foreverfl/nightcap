package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendPostsTheTextAsContent(t *testing.T) {
	var received map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	if err := Send(server.URL, "nightcap blog-v2 · 23:05 → 03:40"); err != nil {
		t.Fatal(err)
	}
	if received["content"] != "nightcap blog-v2 · 23:05 → 03:40" {
		t.Fatalf("content = %q", received["content"])
	}
}

func TestSendWithoutAWebhookSendsNothing(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests++ }))
	defer server.Close()
	if err := Send("", "anything"); err != nil {
		t.Fatal(err)
	}
	if requests != 0 {
		t.Fatalf("want 0 requests, got %d", requests)
	}
}

func TestSlackBodyUsesText(t *testing.T) {
	if got := bodyKey("https://hooks.slack.com/services/T0/B0/x"); got != "text" {
		t.Fatalf("slack key = %q, want text", got)
	}
	if got := bodyKey("https://discord.com/api/webhooks/1/x"); got != "content" {
		t.Fatalf("discord key = %q, want content", got)
	}
}
