# Auth: Bearer → Cookie fallback with CSRF protection

Implementation plan and task list for the dual-credential auth flow.

## Target flow

```
request ──▶ Authorization: Bearer <jwt> present? ──yes──▶ use it, SKIP CSRF
                │
                no
                ▼
            cookie present? ──yes──▶ use it, CSRF check if POST/PUT/PATCH/DELETE
                │                          │
                no                         ▼
                │                    CSRF ok? ──no──▶ 403
                ▼                          │
              401                          yes
                                           ▼
                              verify JWT (sig, exp, iss, aud)
                                           │
                                     fail ──┴── ok
                                      │         │
                                     401        ▼
                                           Principal in context
                                           │
                                           ▼
                                         handler
```

## Chosen approach

Two credentials, one session:

- **Bearer** (`Authorization: Bearer <jwt>`) — used by API clients and the existing
  frontend `localStorage` flow. **No CSRF check**, because a cross-site attacker
  cannot make a browser attach an `Authorization` header.
- **Cookie** (HttpOnly, `SameSite=Lax`, `Secure`) — used by the browser session.
  **CSRF-checked** on `POST`/`PUT`/`PATCH`/`DELETE` via a signed double-submit token:
  a second non-HttpOnly cookie holding `random.HMAC-SHA256(jwtSecret, random)`,
  echoed by the client in the `X-CSRF-TOKEN` header.

The `Authorization` header wins when both are present. No server-side session
store is needed — the JWT remains the only source of truth.

## Current state

Already implemented — no work needed:

| Piece | Location |
|---|---|
| JWT verify: signature, expiry, issuer, audience | `backend/pkg/jwt/jwt.go:80-96` |
| HS256 pinned (`WithValidMethods`), exp required (`WithExpirationRequired`) | `backend/pkg/jwt/jwt.go:87-88` |
| Principal injected into request context | `backend/gateway/internal/adapters/in/http/middleware/auth_middleware.go:34` |
| `GetClaims(ctx)` accessor | `backend/gateway/internal/adapters/in/http/middleware/auth_middleware.go:41-47` |
| Bearer extraction | `backend/gateway/internal/adapters/in/http/response/response.go:109-124` |
| `X-CSRF-Token` allow-listed in CORS (vestigial, nothing validates it) | `backend/gateway/internal/adapters/in/http/server/http_server.go:69` |

Missing — the work:

| Piece | Status |
|---|---|
| Cookie token fallback | DONE |
| CSRF verification on cookie-authed unsafe methods | DONE |
| Cookie issuance on login/register | DONE |
| Logout endpoint that clears cookies | DONE |
| CORS `AllowCredentials` + explicit origins | DONE |
| Frontend `credentials: "include"` + CSRF header | DONE |

## Decisions

- [x] **Whoami reads the principal from context** — done in
      `backend/gateway/internal/adapters/in/http/auth_handler/handler.go:116-138`.
      `WhoamiRequest` now carries `user_id` instead of `token` (proto, `gen/`, gateway
      adapter/DTO, users gRPC handler, whoami usecase all updated).
- [x] **CSRF mechanism — signed double-submit.** Cookie value =
      `random.HMAC-SHA256(jwtSecret, random)`; client echoes the cookie value in
      `X-CSRF-TOKEN`; server verifies the signature, then constant-time compares.
      Stateless, no JWT/proto change, resists subdomain cookie-injection.
- [x] **Logout endpoint — yes.** `POST /api/v0/auth/logout` clears both cookies
      server-side. Without it the cookie lives for the full TTL (10h) and the
      frontend logout stays client-side only.

Rejected alternatives: *plain double-submit* (injectable via a sibling subdomain)
and *CSRF secret as a JWT claim* (would require extending `pkg/jwt.Claims` and
plumbing the claim through the users service — no such change is needed).

## Tasks

### 1. Configuration — done

- [x] Add `CookieConfig` to `backend/pkg/config/default_config.go`:
      `Name`, `CSRFName`, `Path`, `Domain`, `Secure`, `SameSite`, `TTL`.
- [x] Add `Cookie config.CookieConfig` to `backend/gateway/config/config.go`.
- [x] Add a `cookie:` block to `backend/gateway/config/config.yaml`
      (`secure: false` for local http dev, `true` in production).
- [x] Confirm koanf env mapping works. **Caveat:** koanf's env transform replaces
      `_` with `.` (`backend/pkg/config/config.go:40`), so a flat key like
      `csrf_name` becomes unreachable from the env. The struct tags are therefore
      `csrfname` and `samesite` (no underscore) — `COOKIE_CSRFNAME` and
      `COOKIE_SAMESITE` both resolve. Do not rename them back.
- [x] Add the `cookie:` block to `backend/gateway/config/config.example.yaml` —
      `just setup` copies this file, so a fresh clone would otherwise get an
      empty `CookieConfig`.
- [x] Add `config.example.yaml` for both services, a `just setup` recipe, and
      untrack the committed `config.yaml` files (`.gitignore` already had
      `config.yaml`). Removes the secrets from git history going forward.
- [x] Align `cookie.ttl` with `token.ttl` (10h) so the cookie does not outlive
      the JWT it carries.
- [x] Normalize the users service YAML keys to lowercase — koanf lowercases
      keys, so `MinConns` became `minconns` and never matched the
      `min_conns`/`max_conns` tags. The pool size was silently falling back to
      the defaults (20/2) in `backend/pkg/database/postgres.go:28-35`.
- [x] Add the missing `server.host` (gateway) and `grpc.service` (users) keys to
      the example files. `grpc.service` is the name the gRPC health check
      registers (`backend/pkg/grpc_server/grpc_server.go:52`) — it was empty.
- [x] Guard the example files with tests
      (`backend/gateway/config/config_example_test.go`,
      `backend/users/config/config_example_test.go`) backed by the new
      `pkg/config.KoanfKeys` helper. The tests assert, in both directions, that
      the example covers every declared koanf tag and carries no unknown keys.
      This is what catches a stale key such as `csrf_name` vs `csrfname`, which
      otherwise loads as an empty value with no error.

### 2. Token extraction with source

- [x] Change `response.GetTokenFromRequest` (`response.go:119`) to return
      `(token string, source TokenSource, err error)` with
      `TokenSource` ∈ `{SourceBearer, SourceCookie, SourceInvalid}`.
- [ ] Keep the Bearer branch first — an `Authorization` header wins over a cookie.
- [ ] Add cookie lookup using `r.Cookie(...)`; treat a missing cookie as
      "no credential" (not an error).
- [ ] Malformed `Authorization` header → 401 as today, do **not** silently fall
      through to the cookie.

### 3. CSRF verification — signed double-submit
- [x] New package `backend/gateway/internal/adapters/in/http/middleware/csrf` with `Verify` and `Issue`.
- [x] Implement the signed double-submit check: split value/signature, recompute HMAC with the CSRF secret, constant-time compare; also constant-time compare header to cookie value.
- [x] Only enforce for unsafe methods `POST`/`PUT`/`PATCH`/`DELETE`.
- [x] Only enforce when the credential came from the cookie.
- [x] Return 403 on mismatch; missing cookie/header return errors logged by middleware.
- [x] Verify uses `base64.RawURLEncoding` to match `Issue`.

### 4. Auth middleware
- [x] Update `AuthMiddleware` to branch on the token source.
- [x] Cookie source + unsafe method → run the CSRF check before JWT verification.
- [x] Unchanged: JWT verification then principal injection into the context.
- [x] Lower the auth-failure log level from `slog.Error` to `slog.Warn`/`slog.Debug`.

### 5. Cookie issuance
- [x] In `Login` and `Register`, set the auth cookie via `cookies.Set(...)` from `result.Token`.
- [x] Auth cookie: `HttpOnly`, `Secure` (config-driven), `SameSite=Lax`, `Path=/`, `MaxAge` = token TTL.
- [x] CSRF cookie: **not** `HttpOnly` (frontend reads it), `Secure`, `SameSite=Lax`, `Path=/`.
- [x] Cookie `Path` stays `/`.

### 6. Logout endpoint
- [x] Add `POST /auth/logout` in `AuthHandler.Routes()` (outside the protected group).
- [x] Clear both cookies (`MaxAge: -1`, `Expires` epoch).
- [x] Wire the frontend `useLogout` hook to call it.

### 7. CORS
- [x] `AllowCredentials: true`.
- [x] Dev origins explicit; browsers reject `*` together with credentials. In production origins must be listed (enforced by config validation).
- [x] Add `PATCH` to `AllowedMethods`.
- [x] Keep `X-CSRF-Token` in `AllowedHeaders`.

### 8. Frontend

- [x] `frontend/src/lib/api.ts` — read the CSRF cookie and send it as
      `X-CSRF-TOKEN` on unsafe methods.
- [x] `frontend/src/lib/api.ts` — add `credentials: "include"`.
- [x] Switched fully to cookie auth: no `Authorization` header, no `localStorage` token (`token.ts` removed).
- [x] On 401, the session query returns null.
- [x] Frontend calls `POST /auth/logout`.

### 9. API docs

- [ ] Update `backend/gateway/api/openapi.yaml:183-186` — the token description
      mentions only `Authorization: Bearer <token>`; document the cookie and the
      CSRF header.

### 10. Security notes / follow-ups (not blocking)

- [x] The JWT HMAC secret was committed to git in
      `backend/gateway/config/config.yaml` and `backend/users/config/config.yaml`
      (identical shared key). Both files are now untracked and git-ignored, with
      `config.example.yaml` as the template. **Still to do:** rotate the key
      (the old one is still in git history) and load it from a secret manager in
      production.
- [ ] The users service no longer re-verifies the JWT — it trusts the `user_id`
      in the gRPC body. It binds to `localhost` today
      (`backend/users/config/config.yaml`), but consider cross-checking the
      `x-user-id` metadata against `req.user_id` in
      `backend/users/internal/adapters/in/grpc/handler.go:76`.
- [ ] No rate limiting on `POST /api/v0/auth/login` / `/register`.
- [ ] No `WriteTimeout` on the `http.Server`
      (`backend/gateway/internal/adapters/in/http/server/http_server.go:75-80`).
- [ ] `middleware.ClientIPFromHeader("X-Real-IP")`
      (`http_server.go:50`) trusts a client-supplied header.

### 11. Config validation at startup
- [x] Validate `CookieConfig` when the gateway boots (`backend/gateway/cmd/main.go`): `Name` and `CSRFName` non-empty and different, `Path` non-empty, `SameSite` one of `lax`/`strict`/`none`, `TTL` > 0.
- [x] Map `SameSite` via `cookies.sameSite` helper to `http.SameSite` enums.
- [x] Fail fast if `SameSite=none` while `Secure` is false.
- [x] Warn if `cookie.ttl` exceeds `token.ttl`; fail in production if `Secure` is false; fail if `cookie.secret` empty/too short or equals `token.secret` or is the example placeholder. Cookie secret added to example and local configs.

## Verification

- [x] `go build ./...` in all modules — passes.
- [x] `go test` — response/csrf/auth_middleware/config/auth_handler/server pass (and pre-existing tests). Added unit tests for `GetTokenFromRequest`, CSRF, auth middleware.
- [ ] Manual `curl` pass: no header/cookie → 401; cookie only GET → 200; cookie only POST without X-CSRF-Token → 403; cookie only POST with X-CSRF-Token → 200; Authorization: Bearer POST without X-CSRF-Token → 200; tampered CSRF cookie → 403.
- [x] Browser/frontend builds (`npm run build` in frontend passes). Login/logout use cookies/CSRF; frontend sends credentials and CSRF header.
- [ ] Update OpenAPI/docs (README/Auth section).
