package actions

import (
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"

	"yandex_forward_auth/internal/session"
)

func (d *Dependencies) LogoutHandler(c buffalo.Context) error {
	var cookieDomain, sessionID string
	var err error

	cookieDomain, err = d.Config.GetCookieDomain()
	if err != nil {
		return c.Render(http.StatusInternalServerError, nil)
	}

	sessionID, err = session.Parse(c.Request())

	if err == nil && sessionID != "" {
		revokeErr := d.SessionStore.Revoke(c.Request().Context(), sessionID, "logout")
		if revokeErr != nil && !errors.Is(revokeErr, session.ErrNotFound) {
			session.Clear(c.Response(), cookieDomain)
			return c.Render(http.StatusInternalServerError, nil)
		}
	}

	session.Clear(c.Response(), cookieDomain)
	return c.Render(http.StatusNoContent, nil)
}
