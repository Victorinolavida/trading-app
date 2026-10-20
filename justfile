set shell := ["zsh", "-cu"]

mod gateway "backend/gateway"
mod frontend "frontend"
mod users "backend/users"

# Services that own DB migrations (used by the `all` argument)
db_services := "users"

default:
    @just --list

# Run the gateway service
run:
    just gateway run

# Build the gateway binary
build:
    just gateway build

# Run all tests
test:
    just gateway test

# Format all Go code
fmt:
    cd backend && gofmt -w gateway pkg

# Vet all Go code
vet:
    cd backend && go vet ./gateway/... ./pkg/...

# Start postgres + kratos
infra-up:
    docker compose up -d postgres kratos

# Stop infra
infra-down:
    docker compose down

# Tail kratos logs
infra-logs:
    docker compose logs -f kratos

# Apply pending DB migrations: `just migrate users` or `just migrate all`
migrate service:
    just _db migrate {{service}}

# Roll back the last migration batch: `just down users` or `just down all`
down service:
    just _db down {{service}}

# Drop the database and reapply migrations: `just clean users` or `just clean all`
clean service:
    just _db clean {{service}}

# Show migration status: `just status users` or `just status all`
status service:
    just _db status {{service}}

# Regenerate sqlc schema.sql: `just schema users` or `just schema all`
schema service:
    just _db schema {{service}}

[private]
_db cmd service:
    #!/usr/bin/env zsh
    set -eu
    if [[ "{{service}}" == "all" ]]; then
        for s in {{db_services}}; do
            just "$s" "{{cmd}}"
        done
    elif [[ " {{db_services}} " == *" {{service}} "* ]]; then
        just "{{service}}" "{{cmd}}"
    else
        echo "unknown service '{{service}}' (expected: {{db_services}} or all)" >&2
        exit 2
    fi

# Install frontend dependencies
install:
    just frontend install

# Run the frontend dev server (port 8000)
web:
    just frontend run

# Build the frontend
web-build:
    just frontend build

# Run the whole stack: infra + gateway + frontend
dev:
    #!/usr/bin/env zsh
    set -eu
    just infra-up
    killtree() {
        local pid="$1" child
        for child in $(pgrep -P "$pid" 2>/dev/null); do
            killtree "$child"
        done
        kill -TERM "$pid" 2>/dev/null || true
    }
    cleanup() {
        local pid
        for pid in "${stack_pids[@]}"; do
            killtree "$pid"
        done
    }
    trap cleanup EXIT INT TERM
    stack_pids=()
    just run &
    stack_pids+=($!)
    just web &
    stack_pids+=($!)
    wait
