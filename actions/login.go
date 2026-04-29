package actions

import (
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"

	"yandex_forward_auth/internal/oauthstate"
)

func (d *Dependencies) LoginHandler(c buffalo.Context) error {
	returnURL := d.Config.ReturnURLPolicy.Sanitize(c.Request().URL.Query().Get("rd"))

	stateID, err := oauthstate.GenerateStateID()
	if err != nil {
		return c.Render(http.StatusInternalServerError, nil)
	}

	nonce, err := oauthstate.GenerateNonce()
	if err != nil {
		return c.Render(http.StatusInternalServerError, nil)
	}

	codeVerifier, err := oauthstate.GenerateCodeVerifier()
	if err != nil {
		return c.Render(http.StatusInternalServerError, nil)
	}

	now := time.Now().UTC()
	if err := d.OAuthStateStore.Put(c.Request().Context(), stateID, oauthstate.Record{
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		ReturnURL:    returnURL,
		CreatedAt:    now,
		ExpiresAt:    now.Add(oauthstate.DefaultTTL),
	}); err != nil {
		return c.Render(http.StatusInternalServerError, nil)
	}

	oauthstate.SetNonceCookie(c.Response(), nonce, oauthstate.DefaultTTL)

	c.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response().WriteHeader(http.StatusNotImplemented)
	_, err = c.Response().Write([]byte("login oauth transaction created\nstate=" + stateID + "\nreturn_url=" + returnURL + "\n"))

	return err
}
