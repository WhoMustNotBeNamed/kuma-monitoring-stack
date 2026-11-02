#!/usr/bin/env bash

set -euo pipefail

TARGET=${1:-uptime-kuma}

echo "Роняем $TARGET..."
docker exec "$TARGET" sh -c 'kill 1; kill -9 -1' || true

sleep 5

echo "Ждём, пока Docker поднимет контейнер..."
until [ "$(docker inspect -f '{{.State.Health.Status}}' "$TARGET")" = "healthy" ]; do
  sleep 2
done

echo "$TARGET снова healthy."
