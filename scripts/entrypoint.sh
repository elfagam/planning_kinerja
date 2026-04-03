#!/bin/sh
set -e

echo "[ENTRYPOINT] Waiting for database (optional, handled by usecase if needed)..."

# Run migrations automatically
echo "[ENTRYPOINT] Running migrations..."
./migrate

# Start the actual API server
echo "[ENTRYPOINT] Starting e-plan-ai API..."
exec ./main
