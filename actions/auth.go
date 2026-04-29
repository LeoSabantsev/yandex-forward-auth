package actions

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"

	"yandex_forward_auth/internal/session"
)

var authSessionStore session.Store = session.NewMemoryStore()

func AuthHandler(c buffalo.Context) error {
	sessionID, err := session.Parse(c.Request())
	if err != nil || sessionID == "" {
		if err != nil && !errors.Is(err, session.ErrMissingCookie) {
			session.Clear(c.Response())
		}

		http.Redirect(c.Response(), c.Request(), loginRedirectURL(c.Request()), http.StatusFound)
		return nil
	}

	record, err := authSessionStore.Get(c.Request().Context(), sessionID)
	if err != nil {
		session.Clear(c.Response())
		http.Redirect(c.Response(), c.Request(), loginRedirectURL(c.Request()), http.StatusFound)
		return nil
	}

	now := time.Now().UTC()
	if record.Expired(now) || record.Revoked() {
		session.Clear(c.Response())
		http.Redirect(c.Response(), c.Request(), loginRedirectURL(c.Request()), http.StatusFound)
		return nil
	}

	c.Response().Header().Set("X-Auth-User", record.UserID)
	c.Response().Header().Set("X-Auth-Login", record.Login)
	c.Response().Header().Set("X-Auth-Email", record.Email)
	c.Response().Header().Set("X-Auth-Session-ID", sessionID)

	return c.Render(http.StatusNoContent, nil)
}

func loginRedirectURL(r *http.Request) string {
	baseURL := strings.TrimRight(os.Getenv("YA_AUTH_BASE_URL"), "/")
	return baseURL + "/login?rd=" + url.QueryEscape(originalRequestURL(r))
}

func originalRequestURL(r *http.Request) string {
	proto := firstHeaderValue(r.Header.Get("X-Forwarded-Proto"))
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	host := firstHeaderValue(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}

	uri := r.Header.Get("X-Forwarded-Uri")
	if uri == "" {
		uri = r.URL.RequestURI()
	}

	return proto + "://" + host + uri
}

func firstHeaderValue(value string) string {
	if value == "" {
		return ""
	}

	first, _, _ := strings.Cut(value, ",")
	return strings.TrimSpace(first)
}
