# Kubernetes with Traefik

This example protects a demo backend at `/private`. Traefik checks the session
over cluster DNS at `/auth`. The browser reaches the public login and callback
under `/yandex_auth`. Only that public prefix is stripped. The callback has
its own route so it can establish a session before the user returns to the
protected backend.

## Configure

Use a cluster with an amd64 worker for the pinned auth image, Traefik with
the Kubernetes Ingress and CRD providers, the `traefik` ingress class and
the `websecure` entrypoint. Run the commands from this directory.

1. In `resources.yaml`, replace every `app.example.com` with your application's
   DNS name and set `YA_AUTH_ALLOWED_EMAIL_DOMAINS` to the permitted email domain.
   Point the hostname at Traefik.
2. Register a Yandex web OAuth client with access to the user's email address.
   Set its redirect URI to `https://app.example.com/yandex_auth/oauth/callback`,
   using the same hostname as the manifest.
3. Create the namespace with `kubectl create namespace auth-demo`. Provision
   a TLS Secret named `app-tls` there for the hostname. The examples below use
   a certificate trusted by your client.
4. Create the OAuth Secret from a Bash prompt with shell tracing disabled:

   ```bash
   (
     set -euo pipefail
     read -rp 'Client ID: ' client_id
     read -rsp 'Client secret: ' client_secret
     printf '\n'
     test -n "$client_id" && test -n "$client_secret" || exit 1
     printf 'YANDEX_CLIENT_ID=%s\nYANDEX_CLIENT_SECRET=%s\n' "$client_id" "$client_secret" |
       kubectl -n auth-demo create secret generic yandex-oauth --from-env-file=/dev/stdin
   )
   ```

## Deploy and verify

```bash
kubectl -n auth-demo apply -f resources.yaml
kubectl -n auth-demo rollout status deployment/yandex-forward-auth
kubectl -n auth-demo rollout status deployment/private-demo
base=https://app.example.com  # set your actual hostname
curl -sS -o /dev/null -w '%{http_code}\n' "$base/yandex_auth/healthz"
curl -sS -o /dev/null -w '%{http_code} %{redirect_url}\n' "$base/private"
```

Expect HTTP 204 for health and HTTP 302 to the login route with a return URL
for `/private`. Open the protected URL in a fresh browser session. An allowed
user should return to `/private` and see `authenticated`. In a separate
session, a user outside the allowed domain should receive HTTP 403 at the
callback. Keep callback query strings, cookies and client secrets out of
issue reports.
