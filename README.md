# Trading App

A trading application with a React frontend and a Go microservices backend. Search securities, view the order book and market depth, calculate bond prices and yields, and manage orders in the blotter.

## Services

| Service | Role | Local port |
| --- | --- | --- |
| Gateway | HTTP API and authentication | 3001 |
| Users | User accounts and authentication over gRPC | 3030 |
| Securities | Security reference data over gRPC | 3031 |
| Orders | Order creation, matching, and market data over gRPC | 3033 |
| Calculator | Bond price, yield, and cashflow calculations over gRPC | 3034 |
| Orders worker | Expires eligible orders | — |
| Frontend | React application | 8000 |

The services use PostgreSQL; orders also uses Redis and RabbitMQ. The calculator reads securities over gRPC and UDI quotes from the securities database.

## Getting started

Prerequisites: Go 1.26+, Node.js 24+, pnpm, Docker Compose, [just](https://just.systems), and [dbmate](https://github.com/amacneil/dbmate). `buf` and `sqlc` are needed only for code generation.

From the repository root:

```sh
just backend setup               # create local service configs if missing
cp frontend/.env.example frontend/.env
just install                     # install frontend dependencies
docker compose up -d             # start PostgreSQL, Redis, and RabbitMQ
just migrate all                 # create databases and apply migrations
just backend securities seed     # load UDI quotes and securities
```

The Compose Redis instance has no password. Set `redis.password` to an empty string in `backend/orders/config/config.yaml` for this local setup. Check the other local configuration values as needed; the gateway and users token secrets must match. Local `config.yaml` and `.env` files are ignored by Git.

Run each service in a separate terminal:

```sh
just backend run users
just backend run securities
just backend run calculator
just backend run orders
just backend orders run-worker
just backend run gateway
just web
```

Open `http://localhost:8000`. The frontend development server proxies `/api` to the gateway. To stop the infrastructure, run `docker compose down`.

## Securities and calculator

The securities seed imports historical UDI quotes from `backend/securities/internal/adapters/out/postgres/seed_data/udi.csv` without overwriting existing dates, then loads reference and synthetic bonds. The supplied quotes end on **2026-11-19**. Udibono seed rows require a quote dated today or earlier, and the calculator requires a quote for the exact settlement date. Add newer quotes to the CSV or database when needed, then rerun `just backend securities seed`.

The calculator exposes price, yield, and cashflow calculations through `/api/v0/calculator`. The order form can enter a price or yield and preview the other value. Limit orders require an expiration; a separate worker expires eligible orders.

The current calculator infers the coupon rate from the security name, defaults to 5% if it cannot find one, and assumes a 182-day coupon period and 360-day basis. Review these assumptions before relying on calculated values.

## Development commands

Each backend service provides `run`, `build`, `test`, `fmt`, `vet`, and `tidy` recipes. For example:

```sh
just backend test all
just backend build orders         # builds both the orders service and worker
just backend orders build-worker  # builds only the worker
just frontend lint
just frontend build
```

Users, securities, and orders own database migrations. Use a service name or `all`:

```sh
just migrate orders  # apply pending migrations; `just up orders` is equivalent
just down orders     # roll back the last migration batch
just clean orders    # drop the database and reapply migrations
```

The same commands are available through each service's justfile, for example `just backend orders up`. `just backend schema users` regenerates a schema file, and `just backend gen-proto` generates Go protobuf code from `backend/proto`.

## Authentication

The gateway accepts bearer tokens and browser session cookies. Cookie-authenticated state-changing requests require the signed CSRF cookie value in the `X-CSRF-Token` header. Login and registration issue the session and CSRF cookies; logout clears them. Configure cookie and token settings in the gateway and users service configs. The frontend's `VITE_CSRF_COOKIE_NAME` must match the gateway's `cookie.csrfname`.

The gateway API includes authentication, securities, orders, order-book streaming, and calculator endpoints. See `backend/gateway/api/openapi.yaml` for the documented API and [frontend/README.md](frontend/README.md) for frontend details.
