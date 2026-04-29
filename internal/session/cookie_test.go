package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

	Clear(res)

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	require.Equal(t, CookieName, cookie.Name)
	require.Equal(t, "/", cookie.Path)
	require.Equal(t, -1, cookie.MaxAge)
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}
