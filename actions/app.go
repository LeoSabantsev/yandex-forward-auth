package actions

import (
	"sync"
	"yandex_forward_auth/locales"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/middleware/contenttype"
	"github.com/gobuffalo/middleware/forcessl"
	"github.com/gobuffalo/middleware/i18n"
	"github.com/gobuffalo/middleware/paramlogger"
	"github.com/gobuffalo/x/sessions"
	"github.com/rs/cors"
	"github.com/unrolled/secure"

	"yandex_forward_auth/internal/config"
	"yandex_forward_auth/internal/utils/telemetry"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")

var (
	app     *buffalo.App
	appOnce sync.Once
	T       *i18n.Translator
)

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App(cfg config.Config) *buffalo.App {
	appOnce.Do(func() {
		app = newApp(NewDependencies(cfg))
	})

	return app
}

func newApp(deps *Dependencies) *buffalo.App {
	app = buffalo.New(buffalo.Options{
		Env:          ENV,
		SessionStore: sessions.Null{},
		PreWares: []buffalo.PreWare{
			cors.Default().Handler,
		},
		SessionName: "_yandex_forward_auth_session",
	})

	// Automatically redirect to SSL
	// app.Use(forceSSL())

	// Log request parameters (filters apply).
	app.Use(paramlogger.ParameterLogger)

	if deps.Config.Telemetry.Enabled() {
		app.Use(telemetry.Middleware())
	}

	// Set the request content type to JSON
	app.Use(contenttype.Set("application/json"))

	app.GET("/auth", deps.AuthHandler)
	app.GET("/login", deps.LoginHandler)
	app.GET("/oauth/callback", deps.OAuthCallbackHandler)
	app.POST("/logout", deps.LogoutHandler)
	app.GET("/healthz", HealthzHandler)

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(locales.FS(), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}
