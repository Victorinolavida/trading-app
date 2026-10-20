set shell := ["zsh", "-cu"]

mod backend "backend"
mod frontend "frontend"

default:
    @just --list

# Start postgres + kratos
infra-up:
    docker compose up -d postgres kratos

# Stop infra
infra-down:
    docker compose down

# Tail kratos logs
infra-logs:
    docker compose logs -f kratos

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
    just backend run gateway &
    stack_pids+=($!)
    just web &
    stack_pids+=($!)
    wait
