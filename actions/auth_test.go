package actions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"yandex_forward_auth/internal/config"
	"yandex_forward_auth/internal/oauthstate"
	"yandex_forward_auth/internal/session"
	"yandex_forward_auth/internal/yandex"
)

func TestAuthHandler_RedirectsToLoginWhenSessionCookieMissing(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")

	res := performRequest("GET", "http://auth.example.com/auth", nil, nil)

	require.Equal(t, http.StatusFound, res.Code)

	location := res.Header().Get("Location")
	parsed, err := url.Parse(location)
	require.NoError(t, err)

	require.Equal(t, "/login", parsed.Path)
	require.Contains(t, parsed.Query().Get("rd"), "http://auth.example.com/auth")
}

func TestAuthHandler_UsesForwardedHeadersForReturnURL(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")

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

func TestAuthHandler_ClearsCookieAndRedirectsWhenSessionCookieIsInvalid(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")

	reqCookie := &http.Cookie{
		Name:  session.CookieName,
		Value: "test",
	}

	res := performRequest("GET", "http://auth.example.com/auth", nil, []*http.Cookie{reqCookie})

	require.Equal(t, http.StatusFound, res.Code)

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, session.CookieName, cookies[0].Name)
	require.Equal(t, -1, cookies[0].MaxAge)
}

func TestAuthHandler_ClearsCookieAndRedirectsWhenSessionIsMissingFromStore(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")

	sessionID, err := session.Generate()
	require.NoError(t, err)

	withSessionStore(t, session.NewMemoryStore())

	res := performRequest("GET", "http://auth.example.com/auth", nil, []*http.Cookie{
		{Name: session.CookieName, Value: sessionID},
	})

	require.Equal(t, http.StatusFound, res.Code)

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, session.CookieName, cookies[0].Name)
	require.Equal(t, -1, cookies[0].MaxAge)
}

func TestAuthHandler_Returns204AndIdentityHeadersWhenSessionExists(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")
	t.Setenv("YA_AUTH_ALLOWED_USER_IDS", "123456789")

	sessionID, err := session.Generate()
	require.NoError(t, err)

	store := session.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), sessionID, session.Record{
		UserID:    "123456789",
		Login:     "alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}))

	withSessionStore(t, store)

	res := performRequest("GET", "http://auth.example.com/auth", nil, []*http.Cookie{
		{Name: session.CookieName, Value: sessionID},
	})

	require.Equal(t, http.StatusNoContent, res.Code)
	require.Equal(t, "123456789", res.Header().Get("X-Auth-User"))
	require.Equal(t, "alice", res.Header().Get("X-Auth-Login"))
	require.Equal(t, "alice@example.com", res.Header().Get("X-Auth-Email"))
	require.Equal(t, sessionID, res.Header().Get("X-Auth-Session-ID"))
}

func TestAuthHandler_Returns403WhenSessionUserIsNoLongerAllowed(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")
	t.Setenv("YA_AUTH_ALLOWED_USER_IDS", "987654321")

	sessionID, err := session.Generate()
	require.NoError(t, err)

	store := session.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), sessionID, session.Record{
		UserID:    "123456789",
		Login:     "alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}))

	withSessionStore(t, store)

	res := performRequest("GET", "http://auth.example.com/auth", nil, []*http.Cookie{
		{Name: session.CookieName, Value: sessionID},
	})

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Empty(t, res.Header().Get("X-Auth-User"))
}

func TestAuthHandler_AllowsSessionWhenDevAllowAllIsExplicit(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")
	t.Setenv("YA_AUTH_DEV_ALLOW_ALL", "true")

	sessionID, err := session.Generate()
	require.NoError(t, err)

	store := session.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), sessionID, session.Record{
		UserID:    "123456789",
		Login:     "alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}))

	withSessionStore(t, store)

	res := performRequest("GET", "http://auth.example.com/auth", nil, []*http.Cookie{
		{Name: session.CookieName, Value: sessionID},
	})

	require.Equal(t, http.StatusNoContent, res.Code)
	require.Equal(t, "123456789", res.Header().Get("X-Auth-User"))
}

func TestAuthHandler_ClearsCookieAndRedirectsWhenSessionIsExpired(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")

	sessionID, err := session.Generate()
	require.NoError(t, err)

	store := session.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), sessionID, session.Record{
		UserID:    "123456789",
		Login:     "alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
		ExpiresAt: time.Now().UTC().Add(-time.Hour),
	}))

	withSessionStore(t, store)

	res := performRequest("GET", "http://auth.example.com/auth", nil, []*http.Cookie{
		{Name: session.CookieName, Value: sessionID},
	})

	require.Equal(t, http.StatusFound, res.Code)
	require.Len(t, res.Result().Cookies(), 1)
	require.Equal(t, -1, res.Result().Cookies()[0].MaxAge)
}

func TestAuthHandler_ClearsCookieAndRedirectsWhenSessionIsRevoked(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")

	sessionID, err := session.Generate()
	require.NoError(t, err)

	revokedAt := time.Now().UTC()

	store := session.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), sessionID, session.Record{
		UserID:    "123456789",
		Login:     "alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().UTC().Add(-time.Hour),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		RevokedAt: &revokedAt,
	}))

	withSessionStore(t, store)

	res := performRequest("GET", "http://auth.example.com/auth", nil, []*http.Cookie{
		{Name: session.CookieName, Value: sessionID},
	})

	require.Equal(t, http.StatusFound, res.Code)
	require.Len(t, res.Result().Cookies(), 1)
	require.Equal(t, -1, res.Result().Cookies()[0].MaxAge)
}

func TestHealthzHandler_Returns204(t *testing.T) {
	res := performRequest("GET", "http://auth.example.com/healthz", nil, nil)

	require.Equal(t, http.StatusNoContent, res.Code)
}

var testSessionStore session.Store = session.NewMemoryStore()
var testOAuthStateStore oauthstate.Store = oauthstate.NewMemoryStore()
var testYandexClient yandex.Client

func withSessionStore(t *testing.T, store session.Store) {
	t.Helper()

	old := testSessionStore
	testSessionStore = store

	t.Cleanup(func() {
		testSessionStore = old
	})
}

func withOAuthStateStore(t *testing.T, store oauthstate.Store) {
	t.Helper()

	old := testOAuthStateStore
	testOAuthStateStore = store

	t.Cleanup(func() {
		testOAuthStateStore = old
	})
}

func withYandexClient(t *testing.T, client yandex.Client) {
	t.Helper()

	old := testYandexClient
	testYandexClient = client

	t.Cleanup(func() {
		testYandexClient = old
	})
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

	newApp(&Dependencies{
		Config:          config.Load(),
		SessionStore:    testSessionStore,
		OAuthStateStore: testOAuthStateStore,
		YandexClient:    testYandexClient,
	}).ServeHTTP(res, req)

	return res
}
