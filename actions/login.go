package actions

import (
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"

	"yandex_forward_auth/internal/oauthstate"
	"yandex_forward_auth/internal/yandex"
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

	codeChallenge := oauthstate.CodeChallengeS256(codeVerifier)

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

	authURL := yandex.AuthCodeURL(yandex.AuthCodeURLParams{
		ClientID:      d.Config.YandexOAuth.ClientID,
		RedirectURI:   d.Config.BaseURL + "/oauth/callback",
		State:         stateID,
		CodeChallenge: codeChallenge,
	})

	http.Redirect(c.Response(), c.Request(), authURL, http.StatusFound)
	return nil
}
