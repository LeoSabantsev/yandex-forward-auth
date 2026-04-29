package oauthstate

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateStateID(t *testing.T) {
	stateID, err := GenerateStateID()
	require.NoError(t, err)

	require.Len(t, stateID, StateIDLength)
}

func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	require.NoError(t, err)

	require.Len(t, nonce, NonceLength)
	require.NoError(t, ValidateNonce(nonce))
}

func TestValidateNonceRejectsBadValue(t *testing.T) {
	err := ValidateNonce("bad")

	require.ErrorIs(t, err, ErrInvalidNonce)
}

func TestSetAndReadNonceCookie(t *testing.T) {
	nonce, err := GenerateNonce()
	require.NoError(t, err)

	res := httptest.NewRecorder()
	SetNonceCookie(res, nonce, time.Minute)

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, NonceCookieName, cookies[0].Name)
	require.Equal(t, nonce, cookies[0].Value)
	require.Equal(t, "/", cookies[0].Path)
	require.Equal(t, 60, cookies[0].MaxAge)
	require.True(t, cookies[0].HttpOnly)
	require.True(t, cookies[0].Secure)
	require.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)

	req := httptest.NewRequest("GET", "https://auth.example.com/oauth/callback", nil)
	req.AddCookie(cookies[0])

	got, err := ReadNonceCookie(req)
	require.NoError(t, err)
	require.Equal(t, nonce, got)
}

func TestReadNonceCookieMissing(t *testing.T) {
	req := httptest.NewRequest("GET", "https://auth.example.com/oauth/callback", nil)

	_, err := ReadNonceCookie(req)

	require.ErrorIs(t, err, ErrMissingNonceCookie)
}

func TestReadNonceCookieInvalid(t *testing.T) {
	req := httptest.NewRequest("GET", "https://auth.example.com/oauth/callback", nil)
	req.AddCookie(&http.Cookie{Name: NonceCookieName, Value: "bad"})

	_, err := ReadNonceCookie(req)

	require.ErrorIs(t, err, ErrInvalidNonce)
}

func TestClearNonceCookie(t *testing.T) {
	res := httptest.NewRecorder()

	ClearNonceCookie(res)

	cookies := res.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, NonceCookieName, cookies[0].Name)
	require.Equal(t, -1, cookies[0].MaxAge)
}

func TestNonceErrorsAreDistinct(t *testing.T) {
	require.False(t, errors.Is(ErrMissingNonceCookie, ErrInvalidNonce))
}
