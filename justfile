# Root task runner for the trading-app monorepo.
#
# Run `just` (or `just --list`) to see the available recipes. Backend and
# frontend recipes are exposed as modules, so `just backend <recipe>` and
# `just frontend <recipe>` work from the repository root.

set shell := ["zsh", "-cu"]

# =============================================================================
# Modules
# =============================================================================

mod backend "backend"
mod frontend "frontend"

# =============================================================================
# Default
# =============================================================================

# List the recipes available at the repository root.
default:
    @just --list

# =============================================================================
# Local setup
# =============================================================================

# Create local config files from the checked-in examples, if missing.
setup:
    just backend setup

# =============================================================================
# Development
# =============================================================================

# Run the modular monolith (gateway + users + securities + orders in one process).
run:
    just backend run

# Build the modular monolith binary.
build:
    just backend build

# Run every backend test suite (monolith plus each individual module).
test:
    just backend test

# Start the full local stack: infrastructure, monolith backend, and frontend.
dev:
    #!/usr/bin/env zsh
    # Child processes are torn down together on exit. Requires just, docker
    # compose, and local config files created by `just setup`.
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

# =============================================================================
# Code quality
# =============================================================================

# Format Go code.
fmt:
    just backend fmt

# Vet Go code.
vet:
    just backend vet

# Tidy Go dependencies.
tidy:
    just backend tidy

# =============================================================================
# Database
# =============================================================================

# Postgres uses one database (`trading`) with one schema per service.

# Start the local Postgres container.
db-up:
    docker compose up -d postgres

# Apply pending migrations to every service schema.
migrate:
    just --justfile backend/users/justfile migrate
    just --justfile backend/securities/justfile migrate
    just --justfile backend/orders/justfile migrate

# Drop and reapply each service schema.
db-clean:
    just --justfile backend/users/justfile clean
    just --justfile backend/securities/justfile clean
    just --justfile backend/orders/justfile clean

# Roll back the last migration batch for every service schema.
db-down:
    just --justfile backend/users/justfile down
    just --justfile backend/securities/justfile down
    just --justfile backend/orders/justfile down

# Show migration status for every service schema.
db-status:
    just --justfile backend/users/justfile status
    just --justfile backend/securities/justfile status
    just --justfile backend/orders/justfile status

# Regenerate sqlc schema.sql for every service schema.
db-schema:
    just --justfile backend/users/justfile schema
    just --justfile backend/securities/justfile schema
    just --justfile backend/orders/justfile schema

# Seed reference securities. Idempotent.
db-seed:
    just --justfile backend/securities/justfile seed

# Drop, reapply migrations, and seed reference data.
db-reset:
    just db-clean
    just db-seed

# =============================================================================
# Infrastructure
# =============================================================================

# Start the local infrastructure (Postgres, Redis, RabbitMQ) from docker-compose.yaml.
infra-up:
    docker compose up -d postgres redis rabbitmq

# Stop and remove the local infrastructure containers.
infra-down:
    docker compose down

# =============================================================================
# Frontend
# =============================================================================

# Install frontend dependencies.
install:
    just frontend install

# Run the frontend development server on http://localhost:8000.
web:
    just frontend run

# Build the frontend production bundle.
web-build:
    just frontend build

# Legacy microservice lifecycle recipes are commented out in the per-service
# justfiles; the full layout lives on the `microservices` branch.
