package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

func HealthzHandler(c buffalo.Context) error {
	return c.Render(http.StatusNoContent, nil)
}
