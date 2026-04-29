package actions

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"yandex_forward_auth/internal/oauthstate"
)

func TestLoginHandler_CreatesOAuthTransaction(t *testing.T) {
	t.Setenv("YA_AUTH_ALLOWED_RETURN_HOSTS", "app.example.com")
	t.Setenv("YA_AUTH_DEFAULT_REDIRECT_URL", "https://app.example.com/")
	t.Setenv("GO_ENV", "production")

	res := performRequest("GET", "http://auth.example.com/login?rd=https://app.example.com/private", nil, nil)

	require.Equal(t, http.StatusNotImplemented, res.Code)
	require.Contains(t, res.Body.String(), "login oauth transaction created")
	require.Contains(t, res.Body.String(), "return_url=https://app.example.com/private")

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
	t.Setenv("GO_ENV", "production")

	res := performRequest("GET", "http://auth.example.com/login?rd=https://evil.example.com/private", nil, nil)

	require.Equal(t, http.StatusNotImplemented, res.Code)
	require.Contains(t, res.Body.String(), "return_url=https://app.example.com/")
}
