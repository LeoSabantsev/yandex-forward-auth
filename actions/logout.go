package actions

import (
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"

	"yandex_forward_auth/internal/session"
)

func LogoutHandler(c buffalo.Context) error {
	sessionID, err := session.Parse(c.Request())

	if err == nil && sessionID != "" {
		revokeErr := authSessionStore.Revoke(c.Request().Context(), sessionID, "logout")
		if revokeErr != nil && !errors.Is(revokeErr, session.ErrNotFound) {
			session.Clear(c.Response())
			return c.Render(http.StatusInternalServerError, nil)
		}
	}

	session.Clear(c.Response())
	return c.Render(http.StatusNoContent, nil)
}
