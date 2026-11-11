package ghapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenAsksForTheInstallationThenItsToken(t *testing.T) {
	fakeGitHub := http.NewServeMux()
	fakeGitHub.HandleFunc("GET /repos/foreverfl/blog-v2/installation", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":42}`))
	})
	fakeGitHub.HandleFunc("POST /app/installations/42/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"token":"ghs_fake"}`))
	})
	server := httptest.NewServer(fakeGitHub)
	defer server.Close()
	originalURL := apiURL
	apiURL = server.URL
	t.Cleanup(func() { apiURL = originalURL })

	token, err := Token(context.Background(), "a.jwt", "foreverfl/blog-v2")
	if err != nil {
		t.Fatal(err)
	}

	if token != "ghs_fake" {
		t.Fatalf("token = %q, want ghs_fake", token)
	}
}
