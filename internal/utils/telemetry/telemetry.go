package telemetry

import (
	"context"
	"errors"
	"fmt"
	"io"
	standardlog "log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellogglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"

	"yandex_forward_auth/internal/config"
)

const instrumentationName = "yandex_forward_auth/internal/utils/telemetry"

type ShutdownFunc func(context.Context) error

type setupFunc func(context.Context, *resource.Resource, string, string) (ShutdownFunc, error)

var (
	setupTracingProvider = setupTracing
	setupMetricsProvider = setupMetrics
	setupLoggingProvider = setupLogging
)

func Setup(ctx context.Context, cfg config.TelemetryConfig) (ShutdownFunc, error) {
	if !cfg.Enabled() {
		return func(context.Context) error { return nil }, nil
	}

	resAttrs := []attribute.KeyValue{semconv.ServiceName(cfg.ServiceName)}
	for _, attr := range cfg.ResourceAttrs {
		resAttrs = append(resAttrs, attribute.String(attr.Key, attr.Value))
	}

	res, err := resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(resAttrs...),
	)
	if err != nil {
		return func(context.Context) error { return nil }, fmt.Errorf("create telemetry resource: %w", err)
	}

	var shutdowns []ShutdownFunc
	var setupErrs []error

	if cfg.TracesEnabled {
		shutdown, err := setupTracingProvider(ctx, res, cfg.TracesEndpoint, cfg.TracesProtocol)
		if err != nil {
			setupErrs = append(setupErrs, fmt.Errorf("setup tracing: %w", err))
		} else {
			shutdowns = append(shutdowns, shutdown)
		}
	}

	if cfg.MetricsEnabled {
		shutdown, err := setupMetricsProvider(ctx, res, cfg.MetricsEndpoint, cfg.MetricsProtocol)
		if err != nil {
			setupErrs = append(setupErrs, fmt.Errorf("setup metrics: %w", err))
		} else {
			shutdowns = append(shutdowns, shutdown)
		}
	}

	if cfg.LogsEnabled {
		shutdown, err := setupLoggingProvider(ctx, res, cfg.LogsEndpoint, cfg.LogsProtocol)
		if err != nil {
			setupErrs = append(setupErrs, fmt.Errorf("setup logging: %w", err))
		} else {
			shutdowns = append(shutdowns, shutdown)
		}
	}

	if len(shutdowns) > 0 {
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	}

	return combineShutdowns(shutdowns...), errors.Join(setupErrs...)
}

func Middleware() buffalo.MiddlewareFunc {
	meter := otel.Meter(instrumentationName)
	requestCount, _ := meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Total number of HTTP requests handled."),
		metric.WithUnit("{request}"),
	)
	requestDuration, _ := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("Duration of HTTP requests handled."),
		metric.WithUnit("s"),
	)

	tracer := otel.Tracer(instrumentationName)

	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			req := c.Request()
			ctx := otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
			spanName := requestName(c)

			ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
			*req = *req.WithContext(ctx)

			startedAt := time.Now()
			err := next(c)
			duration := time.Since(startedAt).Seconds()
			statusCode := responseStatus(c.Response())

			attrs := []attribute.KeyValue{
				attribute.String("http.request.method", req.Method),
				attribute.String("url.path", req.URL.Path),
				attribute.Int("http.response.status_code", statusCode),
			}
			if route, ok := currentRoute(c); ok {
				attrs = append(attrs, attribute.String("http.route", route.Path))
			}

			span.SetAttributes(attrs...)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else if statusCode >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, http.StatusText(statusCode))
			}
			span.End()

			requestCount.Add(ctx, 1, metric.WithAttributes(attrs...))
			requestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))

			return err
		}
	}
}

func setupTracing(ctx context.Context, res *resource.Resource, endpoint string, protocol string) (ShutdownFunc, error) {
	standardlog.Printf("tracing endpoint: %s", endpoint)
	exporter, err := newTraceExporter(ctx, endpoint, protocol)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)

	return provider.Shutdown, nil
}

func setupMetrics(ctx context.Context, res *resource.Resource, endpoint string, protocol string) (ShutdownFunc, error) {
	standardlog.Printf("metrics endpoint: %s", endpoint)
	exporter, err := newMetricExporter(ctx, endpoint, protocol)
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(provider)

	return provider.Shutdown, nil
}

func setupLogging(ctx context.Context, res *resource.Resource, endpoint string, protocol string) (ShutdownFunc, error) {
	standardlog.Printf("logs endpoint: %s", endpoint)
	exporter, err := newLogExporter(ctx, endpoint, protocol)
	if err != nil {
		return nil, err
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
	otellogglobal.SetLoggerProvider(provider)

	originalLogOutput := standardlog.Writer()
	otelLogWriter := slog.NewLogLogger(
		otelslog.NewHandler(instrumentationName, otelslog.WithLoggerProvider(provider)),
		slog.LevelInfo,
	).Writer()
	standardlog.SetOutput(io.MultiWriter(originalLogOutput, otelLogWriter))

	return func(ctx context.Context) error {
		standardlog.SetOutput(originalLogOutput)
		return provider.Shutdown(ctx)
	}, nil
}

func newTraceExporter(ctx context.Context, endpoint string, protocol string) (sdktrace.SpanExporter, error) {
	if protocol == "http/protobuf" {
		return otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
	}

	return otlptracegrpc.New(ctx, otlptracegrpc.WithEndpointURL(endpoint))
}

func newMetricExporter(ctx context.Context, endpoint string, protocol string) (sdkmetric.Exporter, error) {
	if protocol == "http/protobuf" {
		return otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(endpoint))
	}

	return otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpointURL(endpoint))
}

func newLogExporter(ctx context.Context, endpoint string, protocol string) (sdklog.Exporter, error) {
	if protocol == "http/protobuf" {
		return otlploghttp.New(ctx, otlploghttp.WithEndpointURL(endpoint))
	}

	return otlploggrpc.New(ctx, otlploggrpc.WithEndpointURL(endpoint))
}

func requestName(c buffalo.Context) string {
	if route, ok := currentRoute(c); ok && route.Path != "" {
		return route.Method + " " + route.Path
	}

	req := c.Request()
	return req.Method + " " + req.URL.Path
}

func currentRoute(c buffalo.Context) (buffalo.RouteInfo, bool) {
	route, ok := c.Data()["current_route"].(buffalo.RouteInfo)
	return route, ok
}

func responseStatus(w http.ResponseWriter) int {
	if res, ok := w.(*buffalo.Response); ok && res.Status > 0 {
		return res.Status
	}

	return http.StatusOK
}

func combineShutdowns(shutdowns ...ShutdownFunc) ShutdownFunc {
	return func(ctx context.Context) error {
		var errs []error
		for i := len(shutdowns) - 1; i >= 0; i-- {
			if err := shutdowns[i](ctx); err != nil {
				errs = append(errs, err)
			}
		}

		return errors.Join(errs...)
	}
}
