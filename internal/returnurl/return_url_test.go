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

func TestPolicySanitizeAllowsHTTPSHostFromSimpleWildcard(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"*app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got_1 := policy.Sanitize("https://testing.app.example.com")
	got_2 := policy.Sanitize("https://app.example.com")

	require.Equal(t, "https://testing.app.example.com", got_1)
	require.Equal(t, "https://app.example.com", got_2)
}

func TestPolicySanitizeAllowsHTTPSHostFromComplexWildcard(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"*testing*.app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got_1 := policy.Sanitize("https://testing.app.example.com")
	got_2 := policy.Sanitize("https://ui.testing.app.example.com")
	got_3 := policy.Sanitize("https://testing.inner.app.example.com")
	got_4 := policy.Sanitize("https://ui.testing.inner.app.example.com")

	require.Equal(t, "https://testing.app.example.com", got_1)
	require.Equal(t, "https://app.example.com/", got_2)
	require.Equal(t, "https://testing.inner.app.example.com", got_3)
	require.Equal(t, "https://app.example.com/", got_4)
}

func TestPolicySanitizeDenyHTTPSHostOutsideWildcard(t *testing.T) {
	policy := Policy{
		AllowedHosts: []string{"*.app.example.com"},
		DefaultURL:   "https://app.example.com/",
	}

	got := policy.Sanitize("https://testing.example.com")

	require.Equal(t, "https://app.example.com/", got)
}
