package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

func LoginHandler(c buffalo.Context) error {
	c.Response().WriteHeader(http.StatusNotImplemented)
	_, _ = c.Response().Write([]byte("login is not implemented yet\n"))
	return nil
}
