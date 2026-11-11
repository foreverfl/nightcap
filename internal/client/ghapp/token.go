package ghapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Token returns a token that lets gh act as the App on ownerRepo ("foreverfl/blog-v2"), good for one hour.
// It fails when the App is not installed on ownerRepo or GitHub refuses the jwt.
func Token(ctx context.Context, jwt, ownerRepo string) (string, error) {
	// the App's install number on ownerRepo's account, e.g. 42 for foreverfl; a token is issued per install
	body, err := callGitHub(ctx, http.MethodGet, "/repos/"+ownerRepo+"/installation", jwt)
	if err != nil {
		return "", err
	}
	var installation struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(body, &installation); err != nil {
		return "", err
	}
	// the one-hour token of that install
	body, err = callGitHub(ctx, http.MethodPost, fmt.Sprintf("/app/installations/%d/access_tokens", installation.ID), jwt)
	if err != nil {
		return "", err
	}
	var accessToken struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &accessToken); err != nil {
		return "", err
	}
	return accessToken.Token, nil
}
