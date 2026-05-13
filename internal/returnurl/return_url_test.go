package returnurl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPolicySanitizeAllowsConfiguredHTTPSHost(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got := policy.Sanitize("https://app.example.com/private?x=1")

	require.Equal(t, "https://app.example.com/private?x=1", got)
}

func TestPolicySanitizeRejectsUnknownHost(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got := policy.Sanitize("https://evil.example.com/private")

	require.Equal(t, "https://app.example.com/", got)
}

func TestPolicySanitizeRejectsProtocolRelativeURL(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got := policy.Sanitize("//evil.example.com/private")

	require.Equal(t, "https://app.example.com/", got)
}

func TestPolicySanitizeRejectsHTTPByDefault(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got := policy.Sanitize("http://app.example.com/private")

	require.Equal(t, "https://app.example.com/", got)
}

func TestPolicySanitizeAllowsHTTPWhenExplicitlyConfigured(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"localhost"},
		DefaultURL:   "http://localhost:3000/",
		AllowHTTP:    true,
	}

	got := policy.Sanitize("http://localhost:8080/private")

	require.Equal(t, "http://localhost:8080/private", got)
}

func TestPolicySanitizeAllowsRelativePath(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got := policy.Sanitize("/private?x=1")

	require.Equal(t, "/private?x=1", got)
}

func TestPolicySanitizeAllowsTrimmedRelativePath(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got := policy.Sanitize("  /private?x=1  ")

	require.Equal(t, "/private?x=1", got)
}

func TestPolicySanitizeRejectsMalformedURL(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got := policy.Sanitize("https://%zz")

	require.Equal(t, "https://app.example.com/", got)
}

func TestPolicySanitizeRejectsEmptyAllowedHost(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{""},
		DefaultURL:   "https://app.example.com/",
	}

	got := policy.Sanitize("https://evil.example.com/private")

	require.Equal(t, "https://app.example.com/", got)
}

func TestPolicySanitizeAllowsHTTPSHostFromSimpleWildcard(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"*.example.com"},
		DefaultURL:   "https://example.com/",
	}

	got1 := policy.Sanitize("https://testing.example.com")
	got2 := policy.Sanitize("https://app.example.com")

	require.Equal(t, "https://testing.example.com", got1)
	require.Equal(t, "https://app.example.com", got2)
}

func TestPolicySanitizeAllowsIPv6LiteralHost(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"[::1]"},
		DefaultURL:   "https://example.com/",
	}

	got := policy.Sanitize("https://[::1]/private")

	require.Equal(t, "https://[::1]/private", got)
}

func TestPolicySanitizeDeniesHTTPSHostOutsideWildcard(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"*.app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got := policy.Sanitize("https://testing.example.com")

	require.Equal(t, "https://app.example.com/", got)
}
