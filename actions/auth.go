package actions

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gobuffalo/buffalo"

	"yandex_forward_auth/internal/session"
)

func AuthHandler(c buffalo.Context) error {
	sessionID, err := session.ParseSessionID(c.Request())
	if err != nil || sessionID == "" {
		http.Redirect(c.Response(), c.Request(), loginRedirectURL(c.Request()), http.StatusFound)
		return nil
	}

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
