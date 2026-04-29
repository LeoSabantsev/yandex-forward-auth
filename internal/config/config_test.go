package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadReadsReturnURLPolicy(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", " https://auth.example.com/ ")
	t.Setenv("YA_AUTH_ALLOWED_RETURN_HOSTS", "app.example.com, admin.example.com")
	t.Setenv("YA_AUTH_DEFAULT_REDIRECT_URL", "https://app.example.com/")
	t.Setenv("GO_ENV", "production")

	cfg := Load()

	require.Equal(t, "https://auth.example.com", cfg.BaseURL)
	require.Equal(t, []string{"app.example.com", "admin.example.com"}, cfg.ReturnURLPolicy.AllowedHosts)
	require.Equal(t, "https://app.example.com/", cfg.ReturnURLPolicy.DefaultURL)
	require.False(t, cfg.ReturnURLPolicy.AllowHTTP)
}

func TestLoadAllowsHTTPOutsideProduction(t *testing.T) {
	t.Setenv("GO_ENV", "development")

	cfg := Load()

	require.True(t, cfg.ReturnURLPolicy.AllowHTTP)
}

func TestSplitCSVTrimsAndDropsEmptyItems(t *testing.T) {
	got := splitCSV(" alice@example.com, , bob@example.com ")

	require.Equal(t, []string{"alice@example.com", "bob@example.com"}, got)
}

func TestLoadReadsAllowlist(t *testing.T) {
	t.Setenv("YA_AUTH_ALLOWED_USER_IDS", "123456789,987654321")
	t.Setenv("YA_AUTH_ALLOWED_EMAILS", "alice@example.com,bob@example.com")
	t.Setenv("YA_AUTH_ALLOWED_EMAIL_DOMAINS", "example.com")
	t.Setenv("YA_AUTH_ALLOWED_LOGINS", "alice,bob")
	t.Setenv("YA_AUTH_DEV_ALLOW_ALL", "true")

	cfg := Load()

	require.Equal(t, []string{"123456789", "987654321"}, cfg.Allowlist.UserIDs)
	require.Equal(t, []string{"alice@example.com", "bob@example.com"}, cfg.Allowlist.Emails)
	require.Equal(t, []string{"example.com"}, cfg.Allowlist.EmailDomains)
	require.Equal(t, []string{"alice", "bob"}, cfg.Allowlist.Logins)
	require.True(t, cfg.Allowlist.DevAllowAll)
}
