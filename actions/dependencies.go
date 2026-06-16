package actions

import (
	"yandex_forward_auth/internal/config"
	"yandex_forward_auth/internal/oauthstate"
	"yandex_forward_auth/internal/session"
	"yandex_forward_auth/internal/yandex"
)

type Dependencies struct {
	Config          config.Config
	SessionStore    session.Store
	OAuthStateStore oauthstate.Store
	YandexClient    yandex.Client
}

func NewDependencies(cfg config.Config) *Dependencies {
	return &Dependencies{
		Config:          cfg,
		SessionStore:    session.NewMemoryStore(),
		OAuthStateStore: oauthstate.NewMemoryStore(),
		YandexClient:    yandex.Client{},
	}
}
