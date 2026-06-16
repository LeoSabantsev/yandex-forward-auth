package telemetry

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/resource"

	"yandex_forward_auth/internal/config"
)

func TestSetupDisabledDoesNotStartTelemetry(t *testing.T) {
	started := make([]string, 0)
	restore := replaceSetupProviders(
		recordingSetup("traces", &started, nil),
		recordingSetup("metrics", &started, nil),
		recordingSetup("logs", &started, nil),
	)
	defer restore()

	shutdown, err := Setup(context.Background(), config.TelemetryConfig{})
	require.NoError(t, err)
	require.Empty(t, started)
	require.NoError(t, shutdown(context.Background()))
}

func TestSetupStartsEnabledTelemetryWithSignalEndpoints(t *testing.T) {
	started := make([]string, 0)
	shutdowns := make([]string, 0)
	endpoints := make(map[string]string)
	protocols := make(map[string]string)
	resourceAttrs := make(map[string]string)

	restore := replaceSetupProviders(
		recordingSetupWithShutdown("traces", &started, endpoints, protocols, resourceAttrs, &shutdowns),
		recordingSetupWithShutdown("metrics", &started, endpoints, protocols, nil, &shutdowns),
		recordingSetupWithShutdown("logs", &started, endpoints, protocols, nil, &shutdowns),
	)
	defer restore()

	shutdown, err := Setup(context.Background(), config.TelemetryConfig{
		Endpoint:        "http://otel.example.com:4318",
		TracesEndpoint:  "http://otel.example.com:4318/v1/traces",
		MetricsEndpoint: "http://otel.example.com:4318/v1/metrics",
		LogsEndpoint:    "http://otel.example.com:4318/v1/logs",
		ServiceName:     "test-service",
		ResourceAttrs: []config.TelemetryResourceAttribute{
			{Key: "deployment.environment", Value: "local"},
		},
		TracesProtocol:  "http/protobuf",
		MetricsProtocol: "grpc",
		LogsProtocol:    "http/protobuf",
		TracesEnabled:   true,
		MetricsEnabled:  true,
		LogsEnabled:     true,
	})

	require.NoError(t, err)
	require.ElementsMatch(t, []string{"traces", "metrics", "logs"}, started)
	require.Equal(t, "http://otel.example.com:4318/v1/traces", endpoints["traces"])
	require.Equal(t, "http://otel.example.com:4318/v1/metrics", endpoints["metrics"])
	require.Equal(t, "http://otel.example.com:4318/v1/logs", endpoints["logs"])
	require.Equal(t, "http/protobuf", protocols["traces"])
	require.Equal(t, "grpc", protocols["metrics"])
	require.Equal(t, "http/protobuf", protocols["logs"])
	require.Equal(t, "test-service", resourceAttrs["service.name"])
	require.Equal(t, "local", resourceAttrs["deployment.environment"])

	require.NoError(t, shutdown(context.Background()))
	require.Equal(t, []string{"logs", "metrics", "traces"}, shutdowns)
}

func TestSetupStartsOnlyEnabledSignals(t *testing.T) {
	started := make([]string, 0)
	restore := replaceSetupProviders(
		recordingSetup("traces", &started, nil),
		recordingSetup("metrics", &started, nil),
		recordingSetup("logs", &started, nil),
	)
	defer restore()

	shutdown, err := Setup(context.Background(), config.TelemetryConfig{
		TracesEndpoint: "http://otel.example.com:4318/v1/traces",
		LogsEndpoint:   "http://otel.example.com:4318/v1/logs",
		ServiceName:    "test-service",
		TracesEnabled:  true,
		LogsEnabled:    true,
	})

	require.NoError(t, err)
	require.ElementsMatch(t, []string{"traces", "logs"}, started)
	require.NoError(t, shutdown(context.Background()))
}

func TestSetupReturnsSignalSetupErrorsAndKeepsSuccessfulShutdowns(t *testing.T) {
	started := make([]string, 0)
	shutdowns := make([]string, 0)
	restore := replaceSetupProviders(
		recordingSetup("traces", &started, nil),
		func(context.Context, *resource.Resource, string, string) (ShutdownFunc, error) {
			started = append(started, "metrics")
			return nil, errors.New("metrics failed")
		},
		shutdownRecordingSetup("logs", &shutdowns),
	)
	defer restore()

	shutdown, err := Setup(context.Background(), config.TelemetryConfig{
		TracesEndpoint:  "http://otel.example.com:4318/v1/traces",
		MetricsEndpoint: "http://otel.example.com:4318/v1/metrics",
		LogsEndpoint:    "http://otel.example.com:4318/v1/logs",
		ServiceName:     "test-service",
		TracesEnabled:   true,
		MetricsEnabled:  true,
		LogsEnabled:     true,
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "setup metrics")
	require.ErrorContains(t, err, "metrics failed")
	require.True(t, slices.Contains(started, "traces"))
	require.True(t, slices.Contains(started, "metrics"))

	require.NoError(t, shutdown(context.Background()))
	require.Equal(t, []string{"logs"}, shutdowns)
}

func replaceSetupProviders(tracing, metrics, logging setupFunc) func() {
	originalTracing := setupTracingProvider
	originalMetrics := setupMetricsProvider
	originalLogging := setupLoggingProvider

	setupTracingProvider = tracing
	setupMetricsProvider = metrics
	setupLoggingProvider = logging

	return func() {
		setupTracingProvider = originalTracing
		setupMetricsProvider = originalMetrics
		setupLoggingProvider = originalLogging
	}
}

func recordingSetup(name string, started *[]string, endpoints map[string]string) setupFunc {
	return func(_ context.Context, res *resource.Resource, endpoint string, _ string) (ShutdownFunc, error) {
		if res == nil {
			return nil, errors.New("resource is nil")
		}

		*started = append(*started, name)
		if endpoints != nil {
			endpoints[name] = endpoint
		}

		return func(context.Context) error { return nil }, nil
	}
}

func recordingSetupWithShutdown(name string, started *[]string, endpoints map[string]string, protocols map[string]string, resourceAttrs map[string]string, shutdowns *[]string) setupFunc {
	return func(_ context.Context, res *resource.Resource, endpoint string, protocol string) (ShutdownFunc, error) {
		if res == nil {
			return nil, errors.New("resource is nil")
		}

		*started = append(*started, name)
		if endpoints != nil {
			endpoints[name] = endpoint
		}
		if protocols != nil {
			protocols[name] = protocol
		}
		if resourceAttrs != nil {
			for _, attr := range res.Attributes() {
				resourceAttrs[string(attr.Key)] = attr.Value.AsString()
			}
		}

		return func(context.Context) error {
			*shutdowns = append(*shutdowns, name)
			return nil
		}, nil
	}
}

func shutdownRecordingSetup(name string, shutdowns *[]string) setupFunc {
	return func(_ context.Context, res *resource.Resource, _ string, _ string) (ShutdownFunc, error) {
		if res == nil {
			return nil, errors.New("resource is nil")
		}

		return func(context.Context) error {
			*shutdowns = append(*shutdowns, name)
			return nil
		}, nil
	}
}
