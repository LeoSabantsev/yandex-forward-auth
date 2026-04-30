package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"
)

const (
	CookieName      = "__Secure-yfa_session"
	SessionIDBytes  = 32
	SessionIDLength = 43
)

var (
	ErrMissingCookie  = errors.New("session cookie is missing")
	ErrInvalidSession = errors.New("session id is invalid")
)

func Generate() (string, error) {
	buf := make([]byte, SessionIDBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func Parse(r *http.Request) (string, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", ErrMissingCookie
		}

		return "", err
	}

	if cookie.Value == "" {
		return "", ErrInvalidSession
	}

	if err := Validate(cookie.Value); err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func Validate(id string) error {
	if len(id) != SessionIDLength {
		return ErrInvalidSession
	}

	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return ErrInvalidSession
		}
	}

	return nil
}

func Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func Set(w http.ResponseWriter, sessionID string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
