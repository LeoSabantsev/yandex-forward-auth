package actions

import (
	"yandex_forward_auth/internal/config"
	"yandex_forward_auth/internal/oauthstate"
	"yandex_forward_auth/internal/session"
)

type Dependencies struct {
	Config          config.Config
	SessionStore    session.Store
	OAuthStateStore oauthstate.Store
}

func NewDependencies() *Dependencies {
	return &Dependencies{
		Config:          config.Load(),
		SessionStore:    session.NewMemoryStore(),
		OAuthStateStore: oauthstate.NewMemoryStore(),
	}
}
