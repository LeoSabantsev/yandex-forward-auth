package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateIDCreatesValidSessionID(t *testing.T) {
	id, err := Generate()

	require.NoError(t, err)
	require.Len(t, id, SessionIDLength)
	require.NoError(t, Validate(id))
}

func TestValidateIDRejectsBadValues(t *testing.T) {
	tests := []string{
		"",
		"short",
		"this-value-has-an-invalid-character-!",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",   // 42 chars
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 44 chars
	}

	for _, tt := range tests {
		require.ErrorIs(t, Validate(tt), ErrInvalidSession)
	}
}

func TestClearCookieExpiresSessionCookie(t *testing.T) {
	res := httptest.NewRecorder()
	domain := "example.com"

	Clear(res, domain)

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	require.Equal(t, CookieName, cookie.Name)
	require.Equal(t, "example.com", cookies[0].Domain)
	require.Equal(t, "/", cookie.Path)
	require.Equal(t, -1, cookie.MaxAge)
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestSetCookie(t *testing.T) {
	sessionID, err := Generate()
	require.NoError(t, err)

	res := httptest.NewRecorder()
	domain := "example.com"

	Set(res, sessionID, time.Hour, domain)

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, CookieName, cookies[0].Name)
	require.Equal(t, sessionID, cookies[0].Value)
	require.Equal(t, "example.com", cookies[0].Domain)
	require.Equal(t, "/", cookies[0].Path)
	require.Equal(t, 3600, cookies[0].MaxAge)
	require.True(t, cookies[0].HttpOnly)
	require.True(t, cookies[0].Secure)
	require.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
}
