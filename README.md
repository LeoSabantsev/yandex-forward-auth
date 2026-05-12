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
