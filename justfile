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
# Infrastructure
# =============================================================================

# Start the local Postgres container defined in docker-compose.yaml.
infra-up:
    docker compose up -d postgres

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

# =============================================================================
# Development workflow
# =============================================================================

# Start the full local stack: infrastructure, gateway, and frontend.
dev:
    #!/usr/bin/env zsh
    # Child processes are torn down together on exit. Requires just, docker
    # compose, and local config files created by `just backend setup`.
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
    just backend run gateway &
    stack_pids+=($!)
    just web &
    stack_pids+=($!)
    wait
