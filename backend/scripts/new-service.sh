#!/usr/bin/env zsh
set -eu

if (( $# < 1 || $# > 2 )); then
    print -u2 "usage: new-service.sh <name> [module]"
    exit 2
fi

name=$1
module=${2:-}

if [[ ! "$name" =~ '^[a-z][a-z0-9_-]*$' ]]; then
    print -u2 "invalid service name '$name' (use lowercase letters, digits, '-' or '_')"
    exit 2
fi

script_dir=${0:A:h}
backend_dir=${script_dir:h}
cd "$backend_dir"

if [[ -e "$name" ]]; then
    print -u2 "service directory '$name' already exists"
    exit 1
fi

if grep -q "^[[:space:]]*\./$name\$" go.work; then
    print -u2 "go.work already references './$name'"
    exit 1
fi

if [[ -z "$module" ]]; then
    prefix=$(awk '$1 == "module" { sub("/[^/]+$", "", $2); print $2; exit }' gateway/go.mod)
    prefix=${prefix:-github.com/victorinolavida}
    module="$prefix/$name"
fi

go_version=$(awk '$1 == "go" { print $2; exit }' go.work)
go_version=${go_version:-1.26.5}

mkdir -p "$name/cmd" "$name/config"

cat > "$name/go.mod" <<EOF
module $module

go $go_version

require github.com/victorinolavida/pkg v0.0.0

replace github.com/victorinolavida/pkg => ../pkg
EOF

cat > "$name/.gitignore" <<'EOF'
bin/
EOF

cat > "$name/justfile" <<EOF
# Task runner for the $name service.
#
# The $name service is its own Go module exposing a gRPC API and owning the
# database schema. Migrations are managed with dbmate; generated DB code with
# sqlc.

set shell := ["zsh", "-cu"]

# =============================================================================
# Configuration
# =============================================================================

# Database connection used by dbmate. Override with DATABASE_URL if needed.
db_url := env_var_or_default("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/trading_${name}?sslmode=disable")

# Directory containing dbmate migration files.
migrations := "internal/adapters/out/postgres/migrations"

# Docker Compose service name for the local Postgres container. Used by pg_dump.
pg_container := "trading-postgres"

# Schema file consumed by sqlc, regenerated from the migrated database.
schema_file := "internal/adapters/out/postgres/schema.sql"

# =============================================================================
# Service lifecycle
# =============================================================================

# Run the $name service.
run:
    go run ./cmd

# Build the $name binary into ./bin/$name.
build:
    go build -o bin/$name ./cmd

# Run the test suite, starting from a clean test cache.
test: cache-clean
    go test ./...

# =============================================================================
# Code quality
# =============================================================================

# Format Go code with gofmt.
fmt:
    gofmt -w .

# Vet Go code with the standard vet tool.
vet:
    go vet ./...

# Tidy dependencies with the workspace disabled so go.mod/go.sum stay self-sufficient.
tidy:
    GOWORK=off go mod tidy

# =============================================================================
# Database migrations
# =============================================================================

# Apply all pending DB migrations.
migrate: db-create
    dbmate --url "{{db_url}}" --migrations-dir "{{migrations}}" up

# Regenerate sqlc schema.sql from the migrated database.
schema: migrate
    docker exec {{pg_container}} pg_dump --schema-only --no-owner --no-privileges \\
        --exclude-table=schema_migrations "{{db_url}}" | grep -v '^\\\\' > "{{schema_file}}"

# Roll back the last migration batch.
down: db-create
    dbmate --url "{{db_url}}" --migrations-dir "{{migrations}}" down

# Drop the database and reapply all migrations.
clean: db-create
    dbmate --url "{{db_url}}" --migrations-dir "{{migrations}}" drop
    dbmate --url "{{db_url}}" --migrations-dir "{{migrations}}" up

# Create a new migration file.
new name:
    dbmate --url "{{db_url}}" --migrations-dir "{{migrations}}" new {{name}}

# Show pending/applied migrations.
status: db-create
    dbmate --url "{{db_url}}" --migrations-dir "{{migrations}}" status

# =============================================================================
# Private helpers
# =============================================================================

# Clear the Go test cache so \`just test\` always runs the full suite.
[private]
cache-clean:
    go clean -testcache

# Create the database if it does not exist yet.
[private]
db-create:
    dbmate --url "{{db_url}}" create 2>/dev/null || true
EOF

mkdir -p "$name/internal/adapters/out/postgres/migrations"

cat > "$name/internal/adapters/out/postgres/sqlc.yaml" <<'EOF'
version: "2"
sql:
  - engine: "postgresql"
    queries: "query.sql"
    schema: "schema.sql"
    gen:
      go:
        package: "sqlc"
        out: "sqlc"
        sql_package: "pgx/v5"
        overrides:
          - db_type: "uuid"
            go_type: "github.com/google/uuid.UUID"
EOF

touch "$name/internal/adapters/out/postgres/query.sql" "$name/internal/adapters/out/postgres/schema.sql"

cat > "$name/config/config.go" <<'EOF'
package config

type Server struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type Config struct {
	Env    string `koanf:"env"`
	Server Server `koanf:"server"`
}
EOF

cat > "$name/config/config.example.yaml" <<'EOF'
# Copy this file to config.yaml and adjust the values.
#   cp config/config.example.yaml config/config.yaml
# Every key can be overridden with an env var (SERVER_HOST, SERVER_PORT, ...).

env: development

server:
  host: localhost
  port: 8080
EOF
cp "$name/config/config.example.yaml" "$name/config/config.yaml"

cat > "$name/config/config_example_test.go" <<'EOF'
package config

import (
	"reflect"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	pkgconfig "github.com/victorinolavida/pkg/config"
)

// The example file is what `just setup` copies into config.yaml. A missing or
// mistyped key there silently produces an empty config value, so it is checked
// against the struct definitions on every test run.
func TestExampleConfigCoversEveryField(t *testing.T) {
	k := koanf.New(".")
	if err := k.Load(file.Provider("config.example.yaml"), yaml.Parser()); err != nil {
		t.Fatalf("load example config: %v", err)
	}

	declared := pkgconfig.KoanfKeys(Config{})
	present := make(map[string]bool, len(declared))
	for _, key := range declared {
		present[key] = true
	}

	for _, key := range declared {
		if !k.Exists(key) {
			t.Errorf("config.example.yaml is missing key %q", key)
		}
	}

	for key := range k.All() {
		if !present[key] {
			t.Errorf("config.example.yaml has unknown key %q (check the koanf tags)", key)
		}
	}
}

func TestExampleConfigUnmarshals(t *testing.T) {
	k := koanf.New(".")
	if err := k.Load(file.Provider("config.example.yaml"), yaml.Parser()); err != nil {
		t.Fatalf("load example config: %v", err)
	}

	var conf Config
	if err := k.Unmarshal("", &conf); err != nil {
		t.Fatalf("unmarshal example config: %v", err)
	}

	for _, key := range pkgconfig.KoanfKeys(Config{}) {
		if isEmpty(k.Get(key)) {
			t.Errorf("config.example.yaml leaves %q empty", key)
		}
	}

	if conf.Server.Port <= 0 {
		t.Error("server.port must be positive")
	}
}

func isEmpty(value any) bool {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}
EOF

cat > "$name/cmd/main.go" <<EOF
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	cfg "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/logger"

	"$module/config"
)

const shutdownTimeout = 15 * time.Second

func main() {
	log := logger.NewLogger()

	conf, err := cfg.NewConfig[config.Config]()
	if err != nil {
		logger.Fatal("error loading config", slog.Any("error", err))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              net.JoinHostPort(conf.Server.Host, strconv.Itoa(conf.Server.Port)),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("http server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server error", slog.Any("error", err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("error shutting down http server", slog.Any("error", err))
	}

	log.Info("server shutdown successfully")
}
EOF

tmp=$(mktemp)
awk -v svc="./$name" '
    /^use \(/ { in_use = 1 }
    in_use && /^\)/ { print "\t" svc; in_use = 0 }
    { print }
' go.work > "$tmp"
mv "$tmp" go.work

tmp=$(mktemp)
awk -v svc="$name" '
    /^mod / { last = NR }
    { lines[NR] = $0 }
    END {
        for (i = 1; i <= NR; i++) {
            print lines[i]
            if (i == last) print "mod " svc " \"" svc "\""
        }
    }
' justfile > "$tmp"
mv "$tmp" justfile

tmp=$(mktemp)
awk -v svc="$name" '
    /^services := "/ { sub(/"$/, " " svc "\"") }
    { print }
' justfile > "$tmp"
mv "$tmp" justfile

tmp=$(mktemp)
awk -v svc="$name" '
    /^db_services := "/ { sub(/"$/, " " svc "\"") }
    { print }
' justfile > "$tmp"
mv "$tmp" justfile

gofmt -w "$name"

if ! (cd "$name" && GOWORK=off go mod tidy); then
    print -u2 "warning: go mod tidy failed, run 'just $name tidy' once dependencies are reachable"
fi

print "created service '$name' ($module)"
print "  just $name run"
print "  just $name test"
