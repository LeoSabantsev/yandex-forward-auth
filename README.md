# Yandex Forward Auth

Use the pre-built Docker image:

https://hub.docker.com/r/levsab/yandex-forward-auth

You do not need to build anything locally.

## Minimal Docker Compose

```yaml
services:
  yandex-forward-auth:
    image: levsab/yandex-forward-auth:latest
    restart: unless-stopped
    environment:
      YA_AUTH_BASE_URL: "https://auth.example.com"
      YA_AUTH_COOKIE_DOMAIN: "example.com"
      YA_AUTH_ALLOWED_RETURN_HOSTS: "app.example.com"
      YA_AUTH_DEFAULT_REDIRECT_URL: "https://app.example.com/"
      YA_AUTH_ALLOWED_USER_IDS: "123456789"
      YANDEX_CLIENT_ID: "${YANDEX_CLIENT_ID}"
      YANDEX_CLIENT_SECRET: "${YANDEX_CLIENT_SECRET}"
      YA_AUTH_SESSION_TTL: "8h"
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://127.0.0.1:3000/healthz"]
      interval: 30s
      timeout: 3s
      retries: 3
```

## Environment Variables

| Variable | Purpose |
| --- | --- |
| `YA_AUTH_BASE_URL` | Public URL of this auth service, used for OAuth callback URLs and redirects. |
| `YA_AUTH_COOKIE_DOMAIN` | Top-level domain, which auth service covers; the image defaults to set from `YA_AUTH_BASE_URL` |
| `YA_AUTH_ALLOWED_RETURN_HOSTS` | Comma-separated list of app hosts that `rd` may redirect back to. |
| `YA_AUTH_DEFAULT_REDIRECT_URL` | Fallback redirect URL when `rd` is missing or not allowed. |
| `YANDEX_CLIENT_ID` | Yandex OAuth application client ID. |
| `YANDEX_CLIENT_SECRET` | Yandex OAuth application client secret. |
| `YA_AUTH_ALLOWED_USER_IDS` | Comma-separated allowed stable Yandex user IDs. |
| `YA_AUTH_ALLOWED_EMAILS` | Comma-separated allowed email addresses. |
| `YA_AUTH_ALLOWED_EMAIL_DOMAINS` | Comma-separated allowed email domains. |
| `YA_AUTH_ALLOWED_LOGINS` | Comma-separated allowed Yandex logins. |
| `YA_AUTH_DEV_ALLOW_ALL` | Allows every authenticated Yandex user when set to `true`. |
| `YA_AUTH_SESSION_TTL` | Local session lifetime as a Go duration, for example `8h` or `30m`. |
| `GO_ENV` | Buffalo environment name; the image defaults to `production`. |
| `ADDR` | Listen address; the image defaults to `0.0.0.0`. |
| `PORT` | Listen port; the image defaults to `3000`. |

### Telemetry parameters 
Inherits [basic otlp env variables](https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/)

| Variable | Description |
| --- | --- |
| `OTEL_SERVICE_NAME` | OpenTelemetry service name. Defaults to `yandex-forward-auth`. |
| `OTEL_RESOURCE_ATTRIBUTES` | Comma-separated OpenTelemetry resource attributes, for example `deployment.environment=production,team=platform`. `service.name` is ignored here; use `OTEL_SERVICE_NAME` instead. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Shared OTLP endpoint used as fallback for traces, metrics, and logs. With `http/protobuf`, `/v1/traces`, `/v1/metrics`, or `/v1/logs` is appended per signal. |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | OTLP traces endpoint. Overrides `OTEL_EXPORTER_OTLP_ENDPOINT` for traces. |
| `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | OTLP metrics endpoint. Overrides `OTEL_EXPORTER_OTLP_ENDPOINT` for metrics. |
| `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` | OTLP logs endpoint. Overrides `OTEL_EXPORTER_OTLP_ENDPOINT` for logs. |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | Shared OTLP protocol fallback for all telemetry signals. Use `http/protobuf` for OTLP/HTTP; any other value uses OTLP/gRPC. |
| `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL` | OTLP traces protocol. Overrides `OTEL_EXPORTER_OTLP_PROTOCOL` for traces. |
| `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL` | OTLP metrics protocol. Overrides `OTEL_EXPORTER_OTLP_PROTOCOL` for metrics. |
| `OTEL_EXPORTER_OTLP_LOGS_PROTOCOL` | OTLP logs protocol. Overrides `OTEL_EXPORTER_OTLP_PROTOCOL` for logs. |


By default - if this parameters are not set, service will not send telemetry by OpenTelemetry libs and will use stdout for logging

If `OTEL_EXPORTER_OTLP_ENDPOINT` is set - it will enable all three signals (logs, metrics, traces) to common endpoint, which can handle them all (as described [here](https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/#otel_exporter_otlp_endpoint)). So if `OTEL_EXPORTER_OTLP_ENDPOINT=http://example.com`:
- logs will be send to `http://example.com/v1/logs`
- metrics will be send to `http://example.com/v1/metrics`
- traces will be send to `http://example.com/v1/traces`

If one of signals exporters is set (`OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`, `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` or `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`) - it will directly be set as endpoint for signal, ignoring `OTEL_EXPORTER_OTLP_ENDPOINT`. For example, if `OTEL_EXPORTER_OTLP_ENDPOINT=http://example.com` and `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT=http://another.example.com/logs`:
- logs will be send to `http://another.example.com/logs`
- metrics will be send to `http://example.com/v1/metrics`
- traces will be send to `http://example.com/v1/traces`

If none of `OTEL_EXPORTER_OTLP_ENDPOINT` or `OTEL_EXPORTER_OTLP_<signal>_ENDPOINT` is set - signal will be disabled

## Traefik YAML

```yaml
http:
  routers:
    yandex-forward-auth:
      rule: "Host(`auth.example.com`)"
      service: yandex-forward-auth

    app:
      rule: "Host(`app.example.com`)"
      service: app
      middlewares:
        - yandex-forward-auth

  services:
    yandex-forward-auth:
      loadBalancer:
        servers:
          - url: "http://yandex-forward-auth:3000"

    app:
      loadBalancer:
        servers:
          - url: "http://app:8080"

  middlewares:
    yandex-forward-auth:
      forwardAuth:
        address: "http://yandex-forward-auth:3000/auth"
        preserveLocationHeader: true
        trustForwardHeader: true
        authRequestHeaders:
          - Cookie
          - X-Forwarded-Method
          - X-Forwarded-Proto
          - X-Forwarded-Host
          - X-Forwarded-Uri
          - X-Forwarded-For
        authResponseHeaders:
          - X-Auth-User
          - X-Auth-Login
          - X-Auth-Email
          - X-Auth-Session-ID
        maxBodySize: 1048576
        maxResponseBodySize: 8192
```
