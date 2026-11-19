# trading-app

Monorepo with a Go backend (`backend/`) and a React frontend (`frontend/`).

The backend runs as a **modular monolith**: the HTTP gateway and the users,
securities and orders bounded contexts live in one process and talk to each
other through their public `api` packages instead of gRPC. The previous
microservices layout is preserved on the `microservices` branch.

## Getting started (modular monolith)

```sh
just setup                   # copy every config.example.yaml -> config.yaml (includes app/)
just infra-up                # start Postgres, Redis and RabbitMQ (docker compose)
just migrate                 # apply the users, securities and orders migrations
just run                     # run the whole backend in one process (HTTP, :3001)
just web                     # run the frontend (Vite, :8000)
```

Or run the infrastructure, backend and frontend together with `just dev`, and run
the backend test suites with `just test`.

### Database

Postgres runs one database, `trading`, with one schema per bounded context:
`users`, `securities` and `orders`. Each service justfile sets `search_path` to
its schema, and dbmate keeps a `schema_migrations` table inside that schema. The
old per-service databases (`trading_users`, `trading_securities`,
`trading_orders`) are obsolete; their data was reset and they can be dropped.

```sh
just db-up       # start Postgres only
just migrate     # apply pending migrations to every service schema
just db-clean    # drop and reapply each service schema
just db-reset    # db-clean + seed reference securities
just db-status   # show migration status per schema
just db-schema   # regenerate sqlc schema.sql from each schema
```

The monolith entrypoint is `backend/app`; its config lives in
`backend/app/config/config.yaml` (template: `config.example.yaml`). Every key can
be overridden with an env var (`SERVER_PORT`, `DATABASE_HOST`, `TOKEN_SECRET`, ...).
The token secret must match between the gateway and users modules inside that
single config. The frontend uses the same pattern with `frontend/.env.example` →
`frontend/.env`.

### Legacy microservices layout

The individual services still build, but the backend justfile no longer exposes
their modules or standalone lifecycle recipes. The per-service run/build recipes
are commented out in their own justfiles, and the full microservices layout
lives on the `microservices` branch. Per-service schema commands are still available from each service's justfile:

```sh
just --justfile backend/users/justfile migrate
just --justfile backend/orders/justfile migrate
```

To run a service on its own, uncomment its recipe and invoke it from its
directory (for example, `cd backend/users && just run`).

## Authentication

- **Bearer**: `Authorization: Bearer <jwt>` is accepted and **skips CSRF**.
- **Cookie** (browser): HttpOnly session cookie + non-HttpOnly signed CSRF cookie. For `POST`/`PUT`/`PATCH`/`DELETE` the client must echo the CSRF cookie value in `X-CSRF-Token`. `GET`/`HEAD`/`OPTIONS` are not CSRF-checked.
- **Login/Register**: issue both cookies (JSON still includes the token for API clients).
- **Logout**: `POST /api/v0/auth/logout` clears both cookies (works without a valid token).
- **CORS**: `credentials: include` with explicit `AllowOrigins` in production; `PATCH` allowed.
- **Config**: `token.*` and `cookie.*` in `gateway/config/config.yaml`. Cookie requires `secret` (≥32 bytes), different from `token.secret`, `csrfname` different from `name`, valid `samesite`, `samesite=none` requires `secure=true`. Frontend reads CSRF cookie name from `VITE_CSRF_COOKIE_NAME`.


### Backend

| Tool | Purpose | Install |
| --- | --- | --- |
| [Go](https://go.dev) (1.26+) | Language; multi-module workspace (`backend/go.work`: `app`, `gateway`, `users`, `securities`, `orders`, `pkg`, `gen`) | `brew install go` |
| [just](https://just.systems) | Task runner (root, `backend/`, and per-service justfiles) | `brew install just` |
| [buf](https://buf.build) | Legacy proto lint/codegen (`buf generate` → `backend/gen/`; see the `microservices` branch); uses remote plugins, no local protoc plugins needed | `brew install buf` |
| [sqlc](https://sqlc.dev) | Generates type-safe Go DB code from `query.sql` + `schema.sql` (`sqlc generate` in the service's postgres dir) | `brew install sqlc` |
| [dbmate](https://databasebeard.com/dbmate) | DB migrations (`just migrate`, `db-clean`, `db-status`, ...) | `brew install dbmate` |
| [PostgreSQL 17](https://www.postgresql.org) | Database, via Docker Compose (`just infra-up`) | containerized |
| [Docker](https://www.docker.com) | Runs Postgres, Redis and RabbitMQ | `brew install docker` |

Notable Go libraries: [chi](https://github.com/go-chi/chi) (gateway router), [pgx](https://github.com/jackc/pgx) (Postgres driver), [grpc](https://google.golang.org/grpc) + [protobuf](https://github.com/protocolbuffers/protobuf) (users service), [jwt/v5](https://github.com/golang-jwt/jwt), [koanf](https://github.com/knadh/koanf) (config), [testify](https://github.com/stretchr/testify) (tests).

Gateway: request-id propagation (HTTP headers + gRPC metadata), auth middleware (dual auth: Bearer + HttpOnly cookie with signed double-submit CSRF on unsafe cookie requests), `/api/v0/auth/login`, `/api/v0/auth/register`, `/api/v0/auth/logout`, `/api/v0/auth/whoami`, orders routes (`POST /api/v0/orders`, `PATCH /api/v0/orders/{orderId}/cancel`, `GET /api/v0/orders/active`, `GET /api/v0/orders?order_ids=...`, `GET /api/v0/orders/securities/{securityId}/book` SSE), validation + gRPC error mapping, OpenAPI in `backend/gateway/api/openapi.yaml`. JWT lives in `pkg/jwt` (HS256, issuer/audience/expiry checks). Services pass client IP/user-agent/request-id via gRPC metadata.

Orders: limit orders require an `expiration` (Go duration string, e.g. `"24h"`) and return an `expires_at` timestamp; market orders skip expiration and are killed once matched. The orders service runs an expiration worker that cancels due orders.

Pending: add Prometheus metrics to gateway and users services.

### Frontend

| Tool | Purpose | Install |
| --- | --- | --- |
| [Node.js](https://nodejs.org) (24+) | Runtime | `brew install node` |
| [pnpm](https://pnpm.io) | Package manager | `corepack enable pnpm` |
| [Vite](https://vite.dev) | Dev server + build | via `pnpm install` |
| [React 19](https://react.dev) | UI library | via `pnpm install` |
| [TypeScript](https://www.typescriptlang.org) | Language | via `pnpm install` |
| [Tailwind CSS 4](https://tailwindcss.com) | Styling | via `pnpm install` |
| [TanStack Query](https://tanstack.com/query) | Server-state management | via `pnpm install` |
| [react-hook-form](https://react-hook-form.com) | Forms | via `pnpm install` |

> Note: `frontend/.npmrc` blocks dependency lifecycle scripts (`ignore-scripts=true`).

### Migration / codegen workflow

```sh
just migrate                    # apply all migrations (dbmate)
just db-schema                  # regenerate schema.sql (dbmate + pg_dump)
cd backend/users/internal/adapters/out/postgres && sqlc generate
```
