package config

import (
	"os"
	"strings"

	"yandex_forward_auth/internal/allowlist"
	"yandex_forward_auth/internal/returnurl"
)

type Config struct {
	BaseURL         string
	ReturnURLPolicy returnurl.Policy
	Allowlist       allowlist.List
}

func Load() Config {
	return Config{
		BaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("YA_AUTH_BASE_URL")), "/"),
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
