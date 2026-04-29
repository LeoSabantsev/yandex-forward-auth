package actions

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"yandex_forward_auth/internal/oauthstate"
)

func TestLoginHandler_CreatesOAuthTransaction(t *testing.T) {
	t.Setenv("YA_AUTH_ALLOWED_RETURN_HOSTS", "app.example.com")
	t.Setenv("YA_AUTH_DEFAULT_REDIRECT_URL", "https://app.example.com/")
	t.Setenv("YA_AUTH_BASE_URL", "https://auth.example.com")
	t.Setenv("YANDEX_CLIENT_ID", "client-id")
	t.Setenv("GO_ENV", "production")

	res := performRequest("GET", "http://auth.example.com/login?rd=https://app.example.com/private", nil, nil)

	require.Equal(t, http.StatusFound, res.Code)
	location := res.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)

	require.Equal(t, "https", parsed.Scheme)
	require.Equal(t, "oauth.yandex.com", parsed.Host)
	require.Equal(t, "/authorize", parsed.Path)

	query := parsed.Query()
	require.Equal(t, "code", query.Get("response_type"))
	require.Equal(t, "client-id", query.Get("client_id"))
	require.Equal(t, "https://auth.example.com/oauth/callback", query.Get("redirect_uri"))
	require.NotEmpty(t, query.Get("state"))
	require.NotEmpty(t, query.Get("code_challenge"))
	require.Equal(t, "S256", query.Get("code_challenge_method"))
	require.Empty(t, query.Get("code_verifier"))

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, oauthstate.NonceCookieName, cookies[0].Name)
	require.True(t, cookies[0].HttpOnly)
	require.True(t, cookies[0].Secure)
	require.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
}

func TestLoginHandler_FallsBackWhenReturnURLIsUnsafe(t *testing.T) {
	t.Setenv("YA_AUTH_ALLOWED_RETURN_HOSTS", "app.example.com")
	t.Setenv("YA_AUTH_DEFAULT_REDIRECT_URL", "https://app.example.com/")
	t.Setenv("YA_AUTH_BASE_URL", "https://auth.example.com")
	t.Setenv("YANDEX_CLIENT_ID", "client-id")
	t.Setenv("GO_ENV", "production")

	res := performRequest("GET", "http://auth.example.com/login?rd=https://evil.example.com/private", nil, nil)

	require.Equal(t, http.StatusFound, res.Code)

	location := res.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)

	require.Equal(t, "oauth.yandex.com", parsed.Host)
	require.NotEmpty(t, parsed.Query().Get("state"))
}
