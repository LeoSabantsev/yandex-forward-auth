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
