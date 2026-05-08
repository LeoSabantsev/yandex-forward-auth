package config

import (
	"fmt"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

func (c *Config) GetCookieDomain() (string, error) {
	if c.CookieDomain != "" {
		return c.CookieDomain, nil
	}

	domain, err := parseDomain(c.BaseURL)
	return domain, err
}

func parseDomain(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	host := u.Hostname()
	if host == "" {
		return "", fmt.Errorf("base URL has empty host")
	}

	domain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", err
	}

	return domain, nil
}
