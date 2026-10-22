# trading-app

Monorepo with a Go backend (`backend/`) and a React frontend (`frontend/`).

## Getting started

```sh
just backend setup           # copy config.example.yaml -> config.yaml for gateway + users
just infra-up                # start Postgres (docker compose)
just backend migrate users   # apply migrations (dbmate)
just backend run users       # run the users service (gRPC, :3030)
just backend run gateway     # run the gateway (HTTP, :3001)
just web                     # run the frontend (Vite, :8000)
```

Config files are read from `<service>/config/config.yaml` and every key can be
overridden with an env var (`SERVER_PORT`, `DATABASE_HOST`, `TOKEN_SECRET`, ...).
`config.yaml` is git-ignored: `backend/*/config/config.example.yaml` is the template
to copy, and the token secret must match between gateway and users. The frontend uses
the same pattern with `frontend/.env.example` → `frontend/.env`.

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
| [Go](https://go.dev) (1.26+) | Language; multi-module workspace (`backend/go.work`: `gateway`, `users`, `pkg`) | `brew install go` |
| [just](https://just.systems) | Task runner (root, `backend/`, and per-service justfiles) | `brew install just` |
| [buf](https://buf.build) | Proto lint/codegen (`just backend gen-proto` → `backend/gen/`); uses remote plugins, no local protoc plugins needed | `brew install buf` |
| [sqlc](https://sqlc.dev) | Generates type-safe Go DB code from `query.sql` + `schema.sql` (`sqlc generate` in the service's postgres dir) | `brew install sqlc` |
| [dbmate](https://databasebeard.com/dbmate) | DB migrations (`just backend migrate users`, `down`, `status`, ...) | `brew install dbmate` |
| [PostgreSQL 17](https://www.postgresql.org) | Database, via Docker Compose (`just infra-up`) | containerized |
| [Docker](https://www.docker.com) | Runs Postgres | `brew install docker` |

Notable Go libraries: [chi](https://github.com/go-chi/chi) (gateway router), [pgx](https://github.com/jackc/pgx) (Postgres driver), [grpc](https://google.golang.org/grpc) + [protobuf](https://github.com/protocolbuffers/protobuf) (users service), [jwt/v5](https://github.com/golang-jwt/jwt), [koanf](https://github.com/knadh/koanf) (config), [testify](https://github.com/stretchr/testify) (tests).

Gateway: request-id propagation (HTTP headers + gRPC metadata), auth middleware (dual auth: Bearer + HttpOnly cookie with signed double-submit CSRF on unsafe cookie requests), `/api/v0/auth/login`, `/api/v0/auth/register`, `/api/v0/auth/logout`, `/api/v0/auth/whoami`, validation + gRPC error mapping, OpenAPI in `backend/gateway/api/openapi.yaml`. JWT lives in `pkg/jwt` (HS256, issuer/audience/expiry checks). Services pass client IP/user-agent/request-id via gRPC metadata.

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
just backend migrate users      # apply migrations (dbmate)
just backend schema users       # regenerate schema.sql (dbmate + pg_dump)
cd backend/users/internal/adapters/secondary_adapter/postgres && sqlc generate
just backend gen-proto          # regenerate backend/gen/ from backend/proto (buf)
```
