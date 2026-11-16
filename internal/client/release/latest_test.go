package release

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

// useFakeGitHub points apiURL at a server that answers every request with status and body.
func useFakeGitHub(t *testing.T, status int, body string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(status)
		writer.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	originalURL := apiURL
	apiURL = server.URL
	t.Cleanup(func() { apiURL = originalURL })
}

func TestNewVersionNoticeNamesTheNewerRelease(t *testing.T) {
	useFakeGitHub(t, http.StatusOK, `[{"tag_name":"v0.1.7"},{"tag_name":"v0.1.6"}]`)
	want := []string{"nightcap v0.1.7 is out (you have v0.1.6). Run `nightcap update`."}
	if got := NewVersionNotice(context.Background(), "v0.1.6"); !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNewVersionNoticeIsEmptyOnTheSameVersion(t *testing.T) {
	useFakeGitHub(t, http.StatusOK, `[{"tag_name":"v0.1.6"}]`)
	if got := NewVersionNotice(context.Background(), "v0.1.6"); got != nil {
		t.Fatalf("the same version should print nothing, got %q", got)
	}
}

func TestNewVersionNoticeSkipsADraft(t *testing.T) {
	useFakeGitHub(t, http.StatusOK, `[{"tag_name":"v0.1.7","draft":true},{"tag_name":"v0.1.6"}]`)
	if got := NewVersionNotice(context.Background(), "v0.1.6"); got != nil {
		t.Fatalf("a draft is not out yet, got %q", got)
	}
}

func TestNewVersionNoticeIsEmptyWhenGitHubFails(t *testing.T) {
	useFakeGitHub(t, http.StatusInternalServerError, "")
	if got := NewVersionNotice(context.Background(), "v0.1.6"); got != nil {
		t.Fatalf("a failed request should print nothing, got %q", got)
	}
}
