package actions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"yandex_forward_auth/internal/oauthstate"
	"yandex_forward_auth/internal/session"
	"yandex_forward_auth/internal/yandex"
)

func TestOAuthCallbackHandler_Returns400ForOAuthError(t *testing.T) {
	res := performRequest("GET", "http://auth.example.com/oauth/callback?error=access_denied", nil, nil)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestOAuthCallbackHandler_Returns400WhenStateIsMissing(t *testing.T) {
	res := performRequest("GET", "http://auth.example.com/oauth/callback?code=abc", nil, nil)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestOAuthCallbackHandler_Returns400WhenCodeIsMissing(t *testing.T) {
	res := performRequest("GET", "http://auth.example.com/oauth/callback?state=abc", nil, nil)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestOAuthCallbackHandler_CreatesSessionAndRedirectsForValidStateRecord(t *testing.T) {
	t.Setenv("YANDEX_CLIENT_ID", "client-id")
	t.Setenv("YA_AUTH_ALLOWED_USER_IDS", "123456789")
	t.Setenv("YA_AUTH_SESSION_TTL", "30m")
	t.Setenv("YA_AUTH_BASE_URL", "https://app.example.com/")

	nonce, err := oauthstate.GenerateNonce()
	require.NoError(t, err)

	store := oauthstate.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), "state-id", oauthstate.Record{
		Nonce:        nonce,
		CodeVerifier: "verifier",
		ReturnURL:    "https://app.example.com/",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}))
	withOAuthStateStore(t, store)
	server := withValidYandexServer(t)
	defer server.Close()

	res := performRequest("GET", "http://auth.example.com/oauth/callback?state=state-id&code=def", nil, []*http.Cookie{
		{Name: oauthstate.NonceCookieName, Value: nonce},
	})

	require.Equal(t, http.StatusFound, res.Code)
	require.Equal(t, "https://app.example.com/", res.Header().Get("Location"))

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 2)

	var sessionCookie *http.Cookie
	var nonceCookie *http.Cookie
	for _, cookie := range cookies {
		switch cookie.Name {
		case session.CookieName:
			sessionCookie = cookie
		case oauthstate.NonceCookieName:
			nonceCookie = cookie
		}
	}

	require.NotNil(t, sessionCookie)
	require.NotEmpty(t, sessionCookie.Value)
	require.Equal(t, 1800, sessionCookie.MaxAge)
	require.True(t, sessionCookie.HttpOnly)
	require.True(t, sessionCookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, sessionCookie.SameSite)

	require.NotNil(t, nonceCookie)
	require.Equal(t, -1, nonceCookie.MaxAge)

	storedSession, err := testSessionStore.Get(context.Background(), sessionCookie.Value)
	require.NoError(t, err)
	require.Equal(t, "123456789", storedSession.UserID)
	require.Equal(t, "alice", storedSession.Login)
	require.Equal(t, "alice@example.com", storedSession.Email)
	require.WithinDuration(t, time.Now().UTC().Add(30*time.Minute), storedSession.ExpiresAt, 5*time.Second)
}

func TestOAuthCallbackCreatedSessionPassesAuth(t *testing.T) {
	t.Setenv("YANDEX_CLIENT_ID", "client-id")
	t.Setenv("YA_AUTH_ALLOWED_USER_IDS", "123456789")
	t.Setenv("YA_AUTH_SESSION_TTL", "30m")
	t.Setenv("YA_AUTH_BASE_URL", "https://app.example.com/")

	sessionStore := session.NewMemoryStore()
	withSessionStore(t, sessionStore)

	nonce, err := oauthstate.GenerateNonce()
	require.NoError(t, err)

	stateStore := oauthstate.NewMemoryStore()
	require.NoError(t, stateStore.Put(context.Background(), "state-id", oauthstate.Record{
		Nonce:        nonce,
		CodeVerifier: "verifier",
		ReturnURL:    "https://app.example.com/",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}))
	withOAuthStateStore(t, stateStore)

	server := withValidYandexServer(t)
	defer server.Close()

	callbackRes := performRequest("GET", "http://auth.example.com/oauth/callback?state=state-id&code=def", nil, []*http.Cookie{
		{Name: oauthstate.NonceCookieName, Value: nonce},
	})
	require.Equal(t, http.StatusFound, callbackRes.Code)

	var sessionCookie *http.Cookie
	for _, cookie := range callbackRes.Result().Cookies() {
		if cookie.Name == session.CookieName {
			sessionCookie = cookie
		}
	}
	require.NotNil(t, sessionCookie)

	authRes := performRequest("GET", "http://auth.example.com/auth", nil, []*http.Cookie{sessionCookie})

	require.Equal(t, http.StatusNoContent, authRes.Code)
	require.Equal(t, "123456789", authRes.Header().Get("X-Auth-User"))
	require.Equal(t, "alice", authRes.Header().Get("X-Auth-Login"))
	require.Equal(t, "alice@example.com", authRes.Header().Get("X-Auth-Email"))
	require.Equal(t, sessionCookie.Value, authRes.Header().Get("X-Auth-Session-ID"))
}

func TestOAuthCallbackHandler_Returns400WhenStateRecordIsMissing(t *testing.T) {
	withOAuthStateStore(t, oauthstate.NewMemoryStore())

	res := performRequest("GET", "http://auth.example.com/oauth/callback?state=missing&code=def", nil, nil)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestOAuthCallbackHandler_Returns400WhenStateRecordIsExpired(t *testing.T) {
	store := oauthstate.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), "state-id", oauthstate.Record{
		Nonce:        "nonce",
		CodeVerifier: "verifier",
		ReturnURL:    "https://app.example.com/",
		CreatedAt:    time.Now().UTC().Add(-10 * time.Minute),
		ExpiresAt:    time.Now().UTC().Add(-5 * time.Minute),
	}))
	withOAuthStateStore(t, store)

	res := performRequest("GET", "http://auth.example.com/oauth/callback?state=state-id&code=def", nil, nil)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestOAuthCallbackHandler_Returns400WhenStateRecordIsUsed(t *testing.T) {
	store := oauthstate.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), "state-id", oauthstate.Record{
		Nonce:        "nonce",
		CodeVerifier: "verifier",
		ReturnURL:    "https://app.example.com/",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}))
	require.NoError(t, store.MarkUsed(context.Background(), "state-id"))
	withOAuthStateStore(t, store)

	res := performRequest("GET", "http://auth.example.com/oauth/callback?state=state-id&code=def", nil, nil)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestOAuthCallbackHandler_Returns400WhenNonceCookieIsMissing(t *testing.T) {
	nonce, err := oauthstate.GenerateNonce()
	require.NoError(t, err)

	store := oauthstate.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), "state-id", oauthstate.Record{
		Nonce:        nonce,
		CodeVerifier: "verifier",
		ReturnURL:    "https://app.example.com/",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}))
	withOAuthStateStore(t, store)

	res := performRequest("GET", "http://auth.example.com/oauth/callback?state=state-id&code=def", nil, nil)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestOAuthCallbackHandler_Returns400WhenNonceCookieDoesNotMatch(t *testing.T) {
	storedNonce, err := oauthstate.GenerateNonce()
	require.NoError(t, err)

	cookieNonce, err := oauthstate.GenerateNonce()
	require.NoError(t, err)

	store := oauthstate.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), "state-id", oauthstate.Record{
		Nonce:        storedNonce,
		CodeVerifier: "verifier",
		ReturnURL:    "https://app.example.com/",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}))
	withOAuthStateStore(t, store)

	res := performRequest("GET", "http://auth.example.com/oauth/callback?state=state-id&code=def", nil, []*http.Cookie{
		{Name: oauthstate.NonceCookieName, Value: cookieNonce},
	})

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestOAuthCallbackHandler_Returns503WhenTokenExchangeFails(t *testing.T) {
	t.Setenv("YANDEX_CLIENT_ID", "client-id")
	t.Setenv("YA_AUTH_ALLOWED_USER_IDS", "123456789")

	nonce, err := oauthstate.GenerateNonce()
	require.NoError(t, err)

	store := oauthstate.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), "state-id", oauthstate.Record{
		Nonce:        nonce,
		CodeVerifier: "verifier",
		ReturnURL:    "https://app.example.com/",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}))
	withOAuthStateStore(t, store)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad token", http.StatusBadRequest)
	}))
	defer server.Close()

	withYandexClient(t, yandex.Client{TokenEndpoint: server.URL})

	res := performRequest("GET", "http://auth.example.com/oauth/callback?state=state-id&code=def", nil, []*http.Cookie{
		{Name: oauthstate.NonceCookieName, Value: nonce},
	})

	require.Equal(t, http.StatusServiceUnavailable, res.Code)
}

func TestOAuthCallbackHandler_Returns401WhenClientIDDoesNotMatch(t *testing.T) {
	t.Setenv("YANDEX_CLIENT_ID", "client-id")
	t.Setenv("YA_AUTH_ALLOWED_USER_IDS", "123456789")

	nonce, err := oauthstate.GenerateNonce()
	require.NoError(t, err)

	store := oauthstate.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), "state-id", oauthstate.Record{
		Nonce:        nonce,
		CodeVerifier: "verifier",
		ReturnURL:    "https://app.example.com/",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}))
	withOAuthStateStore(t, store)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			require.NoError(t, json.NewEncoder(w).Encode(yandex.TokenResponse{
				AccessToken: "access-token",
				TokenType:   "bearer",
			}))
		case "/info":
			require.NoError(t, json.NewEncoder(w).Encode(yandex.UserInfo{
				ID:       "123456789",
				Login:    "alice",
				ClientID: "different-client-id",
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	withYandexClient(t, yandex.Client{
		TokenEndpoint:    server.URL + "/token",
		UserInfoEndpoint: server.URL + "/info",
	})

	res := performRequest("GET", "http://auth.example.com/oauth/callback?state=state-id&code=def", nil, []*http.Cookie{
		{Name: oauthstate.NonceCookieName, Value: nonce},
	})

	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestOAuthCallbackHandler_Returns403WhenUserIsNotAllowed(t *testing.T) {
	t.Setenv("YANDEX_CLIENT_ID", "client-id")
	t.Setenv("YA_AUTH_ALLOWED_USER_IDS", "987654321")

	nonce, err := oauthstate.GenerateNonce()
	require.NoError(t, err)

	store := oauthstate.NewMemoryStore()
	require.NoError(t, store.Put(context.Background(), "state-id", oauthstate.Record{
		Nonce:        nonce,
		CodeVerifier: "verifier",
		ReturnURL:    "https://app.example.com/",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}))
	withOAuthStateStore(t, store)

	server := withValidYandexServer(t)
	defer server.Close()

	res := performRequest("GET", "http://auth.example.com/oauth/callback?state=state-id&code=def", nil, []*http.Cookie{
		{Name: oauthstate.NonceCookieName, Value: nonce},
	})

	require.Equal(t, http.StatusForbidden, res.Code)
}

func withValidYandexServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			require.Equal(t, http.MethodPost, r.Method)
			require.NoError(t, r.ParseForm())
			require.Equal(t, "authorization_code", r.Form.Get("grant_type"))
			require.Equal(t, "def", r.Form.Get("code"))
			require.Equal(t, "verifier", r.Form.Get("code_verifier"))

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(yandex.TokenResponse{
				AccessToken: "access-token",
				TokenType:   "bearer",
			}))

		case "/info":
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "OAuth access-token", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(yandex.UserInfo{
				ID:           "123456789",
				Login:        "alice",
				ClientID:     "client-id",
				DefaultEmail: "alice@example.com",
			}))

		default:
			http.NotFound(w, r)
		}
	}))

	withYandexClient(t, yandex.Client{
		TokenEndpoint:    server.URL + "/token",
		UserInfoEndpoint: server.URL + "/info",
	})

	return server
}
