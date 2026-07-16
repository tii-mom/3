#!/usr/bin/env bash

set -Eeuo pipefail

DEPLOY_DIR=${DEPLOY_DIR:-/opt/sub2api-deploy}
MAX_DISK_PERCENT=${MAX_DISK_PERCENT:-85}
MAX_BACKUP_AGE_HOURS=${MAX_BACKUP_AGE_HOURS:-30}
API_URL=${API_URL:-https://api.3api.shop/}
FRONTEND_URL=${FRONTEND_URL:-https://3api.shop/home}

failures=()

for container in sub2api sub2api-postgres sub2api-redis; do
  state=$(docker inspect "$container" --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' 2>/dev/null || true)
  if [[ "$state" != "healthy" && "$state" != "running" ]]; then
    failures+=("$container=$state")
  fi
done

disk_percent=$(df --output=pcent / | tail -1 | tr -dc '0-9')
if (( disk_percent >= MAX_DISK_PERCENT )); then
  failures+=("disk=${disk_percent}%")
fi

if ! curl --fail --silent --show-error --max-time 15 --output /dev/null "$API_URL"; then
  failures+=("api=unreachable")
fi
if ! curl --fail --silent --show-error --max-time 15 --output /dev/null "$FRONTEND_URL"; then
  failures+=("frontend=unreachable")
fi

cd "$DEPLOY_DIR"
set -a
source ./.env
set +a

login_payload=$(jq -n --arg email "$ADMIN_EMAIL" --arg password "$ADMIN_PASSWORD" '{email:$email,password:$password}')
access_token=$(curl --fail --silent --show-error --max-time 15 \
  -H 'Content-Type: application/json' \
  -d "$login_payload" \
  http://127.0.0.1:8085/api/v1/auth/login | jq -r '.data.access_token // .access_token // empty')

if [[ -z "$access_token" ]]; then
  failures+=("backup_check=login_failed")
else
  latest_completed=$(curl --fail --silent --show-error --max-time 15 \
    -H "Authorization: Bearer $access_token" \
    http://127.0.0.1:8085/api/v1/admin/backups | \
    jq -r '(.data.items // .items // []) | map(select(.status == "completed")) | first | .finished_at // .started_at // empty')

  if [[ -z "$latest_completed" ]]; then
    failures+=("backup=missing")
  else
    backup_epoch=$(date -d "$latest_completed" +%s)
    max_age_seconds=$((MAX_BACKUP_AGE_HOURS * 3600))
    if (( $(date +%s) - backup_epoch > max_age_seconds )); then
      failures+=("backup=stale")
    fi
  fi
fi

if (( ${#failures[@]} > 0 )); then
  message="Sub2API health check failed: ${failures[*]}"
  logger -p daemon.err -t sub2api-health "$message"
  printf '%s\n' "$message" >&2
  exit 1
fi

logger -p daemon.info -t sub2api-health "Sub2API health check passed"
printf 'Sub2API health check passed.\n'
