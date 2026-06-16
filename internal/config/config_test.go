package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func init() {
	os.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")
}

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

func TestLoadReadsYandexOAuthConfig(t *testing.T) {
	t.Setenv("YANDEX_CLIENT_ID", " client-id ")
	t.Setenv("YANDEX_CLIENT_SECRET", " client-secret ")

	cfg := Load()

	require.Equal(t, "client-id", cfg.YandexOAuth.ClientID)
	require.Equal(t, "client-secret", cfg.YandexOAuth.ClientSecret)
}

func TestLoadUsesDefaultSessionTTL(t *testing.T) {
	cfg := Load()

	require.Equal(t, DefaultSessionTTL, cfg.SessionTTL)
}

func TestLoadReadsSessionTTL(t *testing.T) {
	t.Setenv("YA_AUTH_SESSION_TTL", "30m")

	cfg := Load()

	require.Equal(t, 30*time.Minute, cfg.SessionTTL)
}

func TestLoadDisablesTelemetryWithoutOTLPEndpoint(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "")

	cfg := Load()

	require.False(t, cfg.Telemetry.Enabled())
	require.Empty(t, cfg.Telemetry.Endpoint)
	require.Empty(t, cfg.Telemetry.TracesEndpoint)
	require.Empty(t, cfg.Telemetry.MetricsEndpoint)
	require.Empty(t, cfg.Telemetry.LogsEndpoint)
	require.Equal(t, DefaultTelemetryServiceName, cfg.Telemetry.ServiceName)
}

func TestLoadUsesSharedOTLPEndpointForAllTelemetrySignals(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel.example.com:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "")

	cfg := Load()

	require.Equal(t, "http://otel.example.com:4318", cfg.Telemetry.Endpoint)
	require.Equal(t, "http://otel.example.com:4318/v1/traces", cfg.Telemetry.TracesEndpoint)
	require.Equal(t, "http://otel.example.com:4318/v1/metrics", cfg.Telemetry.MetricsEndpoint)
	require.Equal(t, "http://otel.example.com:4318/v1/logs", cfg.Telemetry.LogsEndpoint)
	require.True(t, cfg.Telemetry.TracesEnabled)
	require.True(t, cfg.Telemetry.MetricsEnabled)
	require.True(t, cfg.Telemetry.LogsEnabled)
}

func TestLoadAppendsHTTPPathsToSharedOTLPEndpointWithPathPrefix(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel.example.com:4318/collector")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	cfg := Load()

	require.Equal(t, "http://otel.example.com:4318/collector/v1/traces", cfg.Telemetry.TracesEndpoint)
	require.Equal(t, "http://otel.example.com:4318/collector/v1/metrics", cfg.Telemetry.MetricsEndpoint)
	require.Equal(t, "http://otel.example.com:4318/collector/v1/logs", cfg.Telemetry.LogsEndpoint)
}

func TestLoadUsesSignalSpecificOTLPEndpointsWhenSet(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel.example.com:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://otel.example.com:4318/v1/traces")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://otel.example.com:4318/v1/metrics")
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://otel.example.com:4318/v1/logs")

	cfg := Load()

	require.Equal(t, "http://otel.example.com:4318", cfg.Telemetry.Endpoint)
	require.Equal(t, "http://otel.example.com:4318/v1/traces", cfg.Telemetry.TracesEndpoint)
	require.Equal(t, "http://otel.example.com:4318/v1/metrics", cfg.Telemetry.MetricsEndpoint)
	require.Equal(t, "http://otel.example.com:4318/v1/logs", cfg.Telemetry.LogsEndpoint)
	require.True(t, cfg.Telemetry.TracesEnabled)
	require.True(t, cfg.Telemetry.MetricsEnabled)
	require.True(t, cfg.Telemetry.LogsEnabled)
}

func TestLoadEnablesOnlySignalsWithSpecificOTLPEndpoints(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://otel.example.com:4318/v1/metrics")
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://otel.example.com:4318/v1/logs")

	cfg := Load()

	require.Empty(t, cfg.Telemetry.Endpoint)
	require.Empty(t, cfg.Telemetry.TracesEndpoint)
	require.Equal(t, "http://otel.example.com:4318/v1/metrics", cfg.Telemetry.MetricsEndpoint)
	require.Equal(t, "http://otel.example.com:4318/v1/logs", cfg.Telemetry.LogsEndpoint)
	require.False(t, cfg.Telemetry.TracesEnabled)
	require.True(t, cfg.Telemetry.MetricsEnabled)
	require.True(t, cfg.Telemetry.LogsEnabled)
}

func TestLoadReadsSharedOTLPProtocolForAllTelemetrySignals(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel.example.com:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", " HTTP/Protobuf ")

	cfg := Load()

	require.Equal(t, "http/protobuf", cfg.Telemetry.TracesProtocol)
	require.Equal(t, "http/protobuf", cfg.Telemetry.MetricsProtocol)
	require.Equal(t, "http/protobuf", cfg.Telemetry.LogsProtocol)
}

func TestLoadPrefersSignalSpecificOTLPProtocols(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel.example.com:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", " HTTP/PROTOBUF ")
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", "")

	cfg := Load()

	require.Equal(t, "http/protobuf", cfg.Telemetry.TracesProtocol)
	require.Equal(t, "http/protobuf", cfg.Telemetry.MetricsProtocol)
	require.Equal(t, "grpc", cfg.Telemetry.LogsProtocol)
}

func TestLoadReadsOTELResourceAttributes(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel.example.com:4318")
	t.Setenv("OTEL_SERVICE_NAME", "test-service")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "deployment.environment=local,service.name=ignored,team=core%20platform,missing-value")

	cfg := Load()

	require.Equal(t, []TelemetryResourceAttribute{
		{Key: "deployment.environment", Value: "local"},
		{Key: "team", Value: "core platform"},
	}, cfg.Telemetry.ResourceAttrs)
	require.Equal(t, "test-service", cfg.Telemetry.ServiceName)
}

func TestLoadFallsBackForInvalidSessionTTL(t *testing.T) {
	t.Setenv("YA_AUTH_SESSION_TTL", "nope")

	cfg := Load()

	require.Equal(t, DefaultSessionTTL, cfg.SessionTTL)
}

func TestGetCookieDomain_GetDomainFromSpecifiedEnv(t *testing.T) {
	t.Setenv("YA_AUTH_COOKIE_DOMAIN", "example.com")
	t.Setenv("YA_AUTH_BASE_URL", "https://auth.example.com/entry")

	cfg := Load()
	res, err := cfg.GetCookieDomain()
	require.NoError(t, err)

	require.Equal(t, "example.com", res)
}

func TestGetCookieDomain_GetDomainFromSimpleBaseURL(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "https://example.com/entry")

	cfg := Load()
	res, err := cfg.GetCookieDomain()
	require.NoError(t, err)

	require.Equal(t, "example.com", res)
}

func TestGetCookieDomain_GetDomainFromComplexBaseURL(t *testing.T) {
	t.Setenv("YA_AUTH_BASE_URL", "https://app.example.com/entry")

	cfg := Load()
	res, err := cfg.GetCookieDomain()
	require.NoError(t, err)

	require.Equal(t, "example.com", res)
}
