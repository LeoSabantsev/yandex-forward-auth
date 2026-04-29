package returnurl

import (
	"net"
	"net/url"
	"strings"
)

type Policy struct {
	AllowedHosts []string
	DefaultURL   string
	AllowHTTP    bool
}

func (p *Policy) Sanitize(raw string) string {
	defaultURL := strings.TrimSpace(p.DefaultURL)
	if defaultURL == "" {
		defaultURL = "/"
	}

	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil {
		return defaultURL
	}

	if parsed.IsAbs() {
		return p.sanitizeAbsolute(parsed, defaultURL)
	}

	if parsed.Scheme == "" && parsed.Host == "" && strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return parsed.RequestURI()
	}

	return defaultURL
}

func (p Policy) sanitizeAbsolute(parsed *url.URL, defaultURL string) string {
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && !(p.AllowHTTP && scheme == "http") {
		return defaultURL
	}

	host := normalizeHost(parsed.Host)
	if host == "" || !p.hostAllowed(host) {
		return defaultURL
	}

	return parsed.String()
}

func (p Policy) hostAllowed(host string) bool {
	for _, allowed := range p.AllowedHosts {
		if host == normalizeHost(allowed) {
			return true
		}
	}

	return false
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return ""
	}

	if strings.Contains(host, ":") {
		if withoutPort, _, err := net.SplitHostPort(host); err == nil {
			host = withoutPort
		}
	}

	return strings.TrimSuffix(host, ".")
}
