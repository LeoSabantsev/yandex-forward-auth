package main

import (
	"context"
	"log"
	"time"

	"yandex_forward_auth/actions"
	"yandex_forward_auth/internal/config"
	"yandex_forward_auth/internal/utils/telemetry"
)

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {
	appConfig := config.Load()

	shutdownTelemetry, err := telemetry.Setup(context.Background(), appConfig.Telemetry)
	if err != nil {
		log.Printf("telemetry setup failed, continuing with standard logging: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := shutdownTelemetry(ctx); err != nil {
			log.Printf("telemetry shutdown failed: %v", err)
		}
	}()

	app := actions.App(appConfig)
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}

/*
# Notes about `main.go`

## SSL Support

We recommend placing your application behind a proxy, such as
Apache or Nginx and letting them do the SSL heavy lifting
for you. https://gobuffalo.io/en/docs/proxy

## Buffalo Build

When `buffalo build` is run to compile your binary, this `main`
function will be at the heart of that binary. It is expected
that your `main` function will start your application using
the `app.Serve()` method.

*/
