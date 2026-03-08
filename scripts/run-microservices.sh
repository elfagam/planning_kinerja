#!/usr/bin/env sh
set -eu

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

cd "$ROOT_DIR/services/planning-service" && go mod tidy
cd "$ROOT_DIR/services/renja-service" && go mod tidy
cd "$ROOT_DIR/services/performance-service" && go mod tidy
cd "$ROOT_DIR/services/api-gateway" && go mod tidy

echo "Start each service in separate terminal:"
echo "  cd services/planning-service && go run ./cmd/server"
echo "  cd services/renja-service && go run ./cmd/server"
echo "  cd services/performance-service && go run ./cmd/server"
echo "  cd services/api-gateway && go run ./cmd/server"
