package yandex

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthCodeURL(t *testing.T) {
	rawURL := AuthCodeURL(AuthCodeURLParams{
		ClientID:      "client-id",
		RedirectURI:   "https://auth.example.com/oauth/callback",
		State:         "state-id",
		CodeChallenge: "challenge",
	})

	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)

	require.Equal(t, "https", parsed.Scheme)
	require.Equal(t, "oauth.yandex.com", parsed.Host)
	require.Equal(t, "/authorize", parsed.Path)

	query := parsed.Query()
	require.Equal(t, "code", query.Get("response_type"))
	require.Equal(t, "client-id", query.Get("client_id"))
	require.Equal(t, "https://auth.example.com/oauth/callback", query.Get("redirect_uri"))
	require.Equal(t, "state-id", query.Get("state"))
	require.Equal(t, "challenge", query.Get("code_challenge"))
	require.Equal(t, "S256", query.Get("code_challenge_method"))
	require.Empty(t, query.Get("code_verifier"))
}
