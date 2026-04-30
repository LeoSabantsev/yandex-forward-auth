package oauthstate

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"
)

const (
	StateIDBytes = 32
	NonceBytes   = 32

	StateIDLength = 43
	NonceLength   = 43

	NonceCookieName = "__Host-yfa_oauth_nonce"
	DefaultTTL      = 5 * time.Minute
)

var (
	ErrMissingNonceCookie = errors.New("oauth nonce cookie is missing")
	ErrInvalidNonce       = errors.New("oauth nonce is invalid")
)

func GenerateStateID() (string, error) {
	return generateOpaqueValue(StateIDBytes)
}

func GenerateNonce() (string, error) {
	return generateOpaqueValue(NonceBytes)
}

func ValidateNonce(nonce string) error {
	if len(nonce) != NonceLength {
		return ErrInvalidNonce
	}

	for _, r := range nonce {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return ErrInvalidNonce
		}
	}

	return nil
}

func SetNonceCookie(w http.ResponseWriter, nonce string, ttl time.Duration) {
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	http.SetCookie(w, &http.Cookie{
		Name:     NonceCookieName,
		Value:    nonce,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ReadNonceCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(NonceCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", ErrMissingNonceCookie
		}

		return "", err
	}

	if cookie.Value == "" {
		return "", ErrMissingNonceCookie
	}

	if err := ValidateNonce(cookie.Value); err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func ClearNonceCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     NonceCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func generateOpaqueValue(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
