package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

func LoginHandler(c buffalo.Context) error {
	c.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response().WriteHeader(http.StatusNotImplemented)
	_, err := c.Response().Write([]byte("login is not implemented yet\n"))
	return err
}
