package config

import (
	"net/url"
	"os"
	"strings"
	"time"

	"yandex_forward_auth/internal/allowlist"
	"yandex_forward_auth/internal/returnurl"
)

const DefaultSessionTTL = 8 * time.Hour

const DefaultTelemetryServiceName = "yandex-forward-auth"

type YandexOAuthConfig struct {
	ClientID     string
	ClientSecret string
}

type TelemetryConfig struct {
	Endpoint        string
	TracesEndpoint  string
	MetricsEndpoint string
	LogsEndpoint    string
	ServiceName     string
	ResourceAttrs   []TelemetryResourceAttribute
	TracesProtocol  string
	MetricsProtocol string
	LogsProtocol    string
	TracesEnabled   bool
	MetricsEnabled  bool
	LogsEnabled     bool
}

type TelemetryResourceAttribute struct {
	Key   string
	Value string
}

func (c TelemetryConfig) Enabled() bool {
	return c.TracesEnabled || c.MetricsEnabled || c.LogsEnabled
}

type Config struct {
	BaseURL         string
	CookieDomain    string
	ReturnURLPolicy returnurl.Policy
	Allowlist       allowlist.List
	YandexOAuth     YandexOAuthConfig
	SessionTTL      time.Duration
	Telemetry       TelemetryConfig
}

func Load() Config {
	config := Config{
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
		Telemetry:  telemetryConfig(),
	}

	if config.BaseURL == "" {
		panic("Missing Base URL, which is required")
	}

	return config
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

func telemetryConfig() TelemetryConfig {
	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = DefaultTelemetryServiceName
	}
	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))

	tracesProtocol := otlpProtocol("TRACES")
	metricsProtocol := otlpProtocol("METRICS")
	logsProtocol := otlpProtocol("LOGS")

	tracesEndpoint := otlpEndpoint("TRACES", endpoint, tracesProtocol, "/v1/traces")
	metricsEndpoint := otlpEndpoint("METRICS", endpoint, metricsProtocol, "/v1/metrics")
	logsEndpoint := otlpEndpoint("LOGS", endpoint, logsProtocol, "/v1/logs")

	return TelemetryConfig{
		Endpoint:        endpoint,
		TracesEndpoint:  tracesEndpoint,
		MetricsEndpoint: metricsEndpoint,
		LogsEndpoint:    logsEndpoint,
		ServiceName:     serviceName,
		ResourceAttrs:   otelResourceAttributes(),
		TracesProtocol:  tracesProtocol,
		MetricsProtocol: metricsProtocol,
		LogsProtocol:    logsProtocol,
		TracesEnabled:   tracesEndpoint != "",
		MetricsEnabled:  metricsEndpoint != "",
		LogsEnabled:     logsEndpoint != "",
	}
}

func otlpEndpoint(signal string, fallback string, protocol string, httpPath string) string {
	if endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_" + signal + "_ENDPOINT")); endpoint != "" {
		return endpoint
	}

	if protocol == "http/protobuf" {
		return otlpHTTPURL(fallback, httpPath)
	}

	return fallback
}

func otlpHTTPURL(baseURL string, suffix string) string {
	if baseURL == "" {
		return ""
	}

	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return baseURL
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + suffix
	return parsed.String()
}

func otlpProtocol(signal string) string {
	if protocol := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_" + signal + "_PROTOCOL")); protocol != "" {
		return strings.ToLower(protocol)
	}

	return strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")))
}

func otelResourceAttributes() []TelemetryResourceAttribute {
	pairs := splitCSV(os.Getenv("OTEL_RESOURCE_ATTRIBUTES"))
	attrs := make([]TelemetryResourceAttribute, 0, len(pairs))

	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" || key == "service.name" {
			continue
		}

		decoded, err := url.PathUnescape(strings.TrimSpace(value))
		if err != nil {
			decoded = strings.TrimSpace(value)
		}

		attrs = append(attrs, TelemetryResourceAttribute{
			Key:   key,
			Value: decoded,
		})
	}

	return attrs
}
