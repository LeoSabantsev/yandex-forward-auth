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
		if matchHostPattern(normalizeHost(allowed), host) {
			return true
		}
	}

	return false
}

func matchHostPattern(pattern, host string) bool {
	if pattern == "" {
		return false
	}
	if !strings.Contains(pattern, "*") {
		return pattern == host
	}

	parts := strings.Split(pattern, "*")
	position := 0

	first := parts[0]
	if first != "" {
		if !strings.HasPrefix(host, first) {
			return false
		}
		position = len(first)
	}

	for _, part := range parts[1 : len(parts)-1] {
		if part == "" {
			continue
		}

		index := strings.Index(host[position:], part)
		if index == -1 {
			return false
		}
		position += index + len(part)
	}

	last := parts[len(parts)-1]
	if last == "" {
		return true
	}

	return len(host)-len(last) >= position && strings.HasSuffix(host, last)
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
