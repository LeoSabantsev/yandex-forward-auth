package actions

import (
	"errors"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"

	"yandex_forward_auth/internal/allowlist"
	"yandex_forward_auth/internal/oauthstate"
	"yandex_forward_auth/internal/session"
	"yandex_forward_auth/internal/yandex"
)

func (d *Dependencies) OAuthCallbackHandler(c buffalo.Context) error {
	query := c.Request().URL.Query()

	if query.Get("error") != "" {
		return c.Render(http.StatusBadRequest, nil)
	}

	stateID := query.Get("state")
	if stateID == "" || query.Get("code") == "" {
		return c.Render(http.StatusBadRequest, nil)
	}

	record, err := d.OAuthStateStore.Get(c.Request().Context(), stateID)
	if err != nil {
		if errors.Is(err, oauthstate.ErrNotFound) {
			return c.Render(http.StatusBadRequest, nil)
		}

		return c.Render(http.StatusInternalServerError, nil)
	}

	now := time.Now().UTC()
	if record.Expired(now) || record.Used() {
		return c.Render(http.StatusBadRequest, nil)
	}

	nonce, err := oauthstate.ReadNonceCookie(c.Request())
	if err != nil {
		return c.Render(http.StatusBadRequest, nil)
	}

	if nonce != record.Nonce {
		return c.Render(http.StatusBadRequest, nil)
	}

	if err := d.OAuthStateStore.MarkUsed(c.Request().Context(), stateID); err != nil {
		return c.Render(http.StatusInternalServerError, nil)
	}

	oauthstate.ClearNonceCookie(c.Response())

	token, err := d.YandexClient.ExchangeCode(c.Request().Context(), yandex.TokenRequest{
		ClientID:     d.Config.YandexOAuth.ClientID,
		ClientSecret: d.Config.YandexOAuth.ClientSecret,
		Code:         query.Get("code"),
		CodeVerifier: record.CodeVerifier,
	})
	if err != nil {
		return c.Render(http.StatusServiceUnavailable, nil)
	}

	userInfo, err := d.YandexClient.UserInfo(c.Request().Context(), token.AccessToken)
	if err != nil {
		return c.Render(http.StatusServiceUnavailable, nil)
	}

	if userInfo.ClientID != d.Config.YandexOAuth.ClientID || userInfo.ID == "" {
		return c.Render(http.StatusUnauthorized, nil)
	}

	if !d.Config.Allowlist.Allows(allowlist.Subject{
		UserID: userInfo.ID,
		Email:  userInfo.Email(),
		Login:  userInfo.Login,
	}) {
		return c.Render(http.StatusForbidden, nil)
	}

	sessionID, err := session.Generate()
	if err != nil {
		return c.Render(http.StatusInternalServerError, nil)
	}

	now = time.Now().UTC()
	if err := d.SessionStore.Put(c.Request().Context(), sessionID, session.Record{
		UserID:     userInfo.ID,
		Login:      userInfo.Login,
		Email:      userInfo.Email(),
		CreatedAt:  now,
		ExpiresAt:  now.Add(d.Config.SessionTTL),
		LastSeenAt: now,
	}); err != nil {
		return c.Render(http.StatusInternalServerError, nil)
	}

	session.Set(c.Response(), sessionID, d.Config.SessionTTL)
	http.Redirect(c.Response(), c.Request(), record.ReturnURL, http.StatusFound)
	return nil
}
