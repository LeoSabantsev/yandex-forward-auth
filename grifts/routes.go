package grifts

import (
	"fmt"
	"os"
	"text/tabwriter"

	"yandex_forward_auth/actions"

	"github.com/gobuffalo/grift/grift"
)

var _ = grift.Add("routes", func(c *grift.Context) error {
	app := actions.App()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "METHOD\tPATH\tNAME\tHANDLER")

	for _, r := range app.Routes() {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Method, r.Path, r.PathName, r.HandlerName)
	}

	return w.Flush()
})
