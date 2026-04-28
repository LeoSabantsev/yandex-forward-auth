package actions

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"yandex_forward_auth/internal/session"
)

func TestAuthHandler_RedirectsToLoginWhenSessionCookieMissing(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "")

	res := performRequest("GET", "http://auth.example.com/auth", nil, nil)

	require.Equal(t, http.StatusFound, res.Code)

	location := res.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)

	require.Equal(t, "/login", parsed.Path)
	require.Equal(t, "http://auth.example.com/auth/", parsed.Query().Get("rd"))
}

func TestAuthHandler_UsesForwardedHeadersForReturnURL(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "")

	headers := map[string]string{
		"X-Forwarded-Proto": "https",
		"X-Forwarded-Host":  "app.example.com",
		"X-Forwarded-Uri":   "/private/path?foo=bar",
	}

	res := performRequest("GET", "http://auth.example.com/auth", headers, nil)

	require.Equal(t, http.StatusFound, res.Code)

	location := res.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)

	require.Equal(t, "/login", parsed.Path)
	require.Equal(t, "https://app.example.com/private/path?foo=bar", parsed.Query().Get("rd"))
}

func TestAuthHandler_Returns204WhenSessionCookieExists(t *testing.T) {
	reqCookie := &http.Cookie{
		Name:  session.CookieName,
		Value: "temporary-session-id",
	}

	res := performRequest("GET", "http://auth.example.com/auth", nil, []*http.Cookie{reqCookie})

	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestHealthzHandler_Returns204(t *testing.T) {
	res := performRequest("GET", "http://auth.example.com/healthz", nil, nil)

	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestLoginHandler_IsPlaceholder(t *testing.T) {
	res := performRequest("GET", "http://auth.example.com/login", nil, nil)

	require.Equal(t, http.StatusNotImplemented, res.Code)
	require.Contains(t, res.Body.String(), "login is not implemented yet")
}

func performRequest(method string, target string, headers map[string]string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	res := httptest.NewRecorder()
	App().ServeHTTP(res, req)

	return res
}
