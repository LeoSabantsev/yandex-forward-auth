package session

import (
	"errors"
	"net/http"
)

const CookieName = "__Secure-yfa_session"

var ErrMissingCookie = errors.New("session cookie is missing")

func ParseSessionID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", ErrMissingCookie
		}

		return "", err
	}

	if cookie.Value == "" {
		return "", ErrMissingCookie
	}

	return cookie.Value, nil
}
