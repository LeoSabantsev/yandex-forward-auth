package actions

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"yandex_forward_auth/internal/session"
)

func TestLogoutHandler_Returns204AndClearsCookieWhenCookieMissing(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")

	withSessionStore(t, session.NewMemoryStore())

	res := performRequest("POST", "http://auth.example.com/logout", nil, nil)

	require.Equal(t, http.StatusNoContent, res.Code)

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, session.CookieName, cookies[0].Name)
	require.Equal(t, -1, cookies[0].MaxAge)
}

func TestLogoutHandler_Returns204AndClearsCookieWhenCookieInvalid(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")

	withSessionStore(t, session.NewMemoryStore())

	res := performRequest("POST", "http://auth.example.com/logout", nil, []*http.Cookie{
		{Name: session.CookieName, Value: "bad"},
	})

	require.Equal(t, http.StatusNoContent, res.Code)

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, session.CookieName, cookies[0].Name)
	require.Equal(t, -1, cookies[0].MaxAge)
}

func TestLogoutHandler_RevokesSessionAndClearsCookie(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")

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

	res := performRequest("POST", "http://auth.example.com/logout", nil, []*http.Cookie{
		{Name: session.CookieName, Value: sessionID},
	})

	require.Equal(t, http.StatusNoContent, res.Code)

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, session.CookieName, cookies[0].Name)
	require.Equal(t, -1, cookies[0].MaxAge)

	record, err := store.Get(context.Background(), sessionID)
	require.NoError(t, err)
	require.True(t, record.Revoked())
	require.Equal(t, "logout", record.RevokedReason)
}
