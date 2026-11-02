#!/bin/sh
# Накатывает datasource и дашборд в Grafana через API
set -eu

GRAFANA_URL=${GRAFANA_URL:-http://localhost:3000}
AUTH="$GF_ADMIN_USER:$GF_ADMIN_PASSWORD"

until curl -sf "$GRAFANA_URL/api/health" >/dev/null; do sleep 2; done

curl -s -u "$AUTH" -H "Content-Type: application/json" \
  -X POST "$GRAFANA_URL/api/datasources" \
  -d @monitoring/grafana/datasources/prometheus.json


curl -sf -u "$AUTH" -H "Content-Type: application/json" \
  -X POST "$GRAFANA_URL/api/dashboards/db" \
  -d "{\"dashboard\": $(cat monitoring/grafana/dashboards/uptime-kuma-health.json), \"overwrite\": true}"
  