package config

import (
	"os"
	"strings"
	"time"

	"yandex_forward_auth/internal/allowlist"
	"yandex_forward_auth/internal/returnurl"
)

const DefaultSessionTTL = 8 * time.Hour

type YandexOAuthConfig struct {
	ClientID     string
	ClientSecret string
}

type Config struct {
	BaseURL         string
	CookieDomain    string
	ReturnURLPolicy returnurl.Policy
	Allowlist       allowlist.List
	YandexOAuth     YandexOAuthConfig
	SessionTTL      time.Duration
}

func Load() Config {
	return Config{
		BaseURL:      strings.TrimRight(strings.TrimSpace(os.Getenv("YA_AUTH_BASE_URL")), "/"),
		CookieDomain: strings.TrimSpace(os.Getenv("YA_AUTH_COOKIE_DOMAIN")),
		ReturnURLPolicy: returnurl.Policy{
			AllowedHosts: splitCSV(os.Getenv("YA_AUTH_ALLOWED_RETURN_HOSTS")),
			DefaultURL:   strings.TrimSpace(os.Getenv("YA_AUTH_DEFAULT_REDIRECT_URL")),
			AllowHTTP:    os.Getenv("GO_ENV") != "production",
		},
		Allowlist: allowlist.List{
			UserIDs:      splitCSV(os.Getenv("YA_AUTH_ALLOWED_USER_IDS")),
			Emails:       splitCSV(os.Getenv("YA_AUTH_ALLOWED_EMAILS")),
			EmailDomains: splitCSV(os.Getenv("YA_AUTH_ALLOWED_EMAIL_DOMAINS")),
			Logins:       splitCSV(os.Getenv("YA_AUTH_ALLOWED_LOGINS")),
			DevAllowAll:  envBool("YA_AUTH_DEV_ALLOW_ALL"),
		},
		YandexOAuth: YandexOAuthConfig{
			ClientID:     strings.TrimSpace(os.Getenv("YANDEX_CLIENT_ID")),
			ClientSecret: strings.TrimSpace(os.Getenv("YANDEX_CLIENT_SECRET")),
		},
		SessionTTL: sessionTTL(),
	}
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}

	return items
}

func envBool(name string) bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(name)), "true")
}

func sessionTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv("YA_AUTH_SESSION_TTL"))
	if raw == "" {
		return DefaultSessionTTL
	}

	ttl, err := time.ParseDuration(raw)
	if err != nil || ttl <= 0 {
		return DefaultSessionTTL
	}

	return ttl
}
