#!/usr/bin/env bash

set -Eeuo pipefail
umask 077

DEPLOY_DIR=${DEPLOY_DIR:-/opt/sub2api-deploy}
ENV_FILE=${ENV_FILE:-$DEPLOY_DIR/.env}
BACKUP_DIR=${BACKUP_DIR:-$DEPLOY_DIR/backups/preflight}
IMAGE_REPOSITORY=${IMAGE_REPOSITORY:-ghcr.io/tii-mom/3}
IMAGE_DIGEST=${1:-}
CONFIRMATION=${2:-}
PROD_POSTGRES_CONTAINER=${PROD_POSTGRES_CONTAINER:-sub2api-postgres}
PROD_APP_CONTAINER=${PROD_APP_CONTAINER:-sub2api}
RESTORE_IMAGE=${RESTORE_IMAGE:-postgres:18-alpine}
RESTORE_TIMEOUT=${RESTORE_TIMEOUT:-120}

file_mode() {
  stat -c '%a' "$1" 2>/dev/null || stat -f '%Lp' "$1"
}

available_bytes_for() {
  if df --output=avail -B1 "$1" >/dev/null 2>&1; then
    df --output=avail -B1 "$1" | tail -1 | tr -d ' '
  else
    df -Pk "$1" | awk 'NR == 2 { print $4 * 1024 }'
  fi
}

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

if [[ ! "$IMAGE_DIGEST" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  printf 'Usage: %s <sha256:image-digest> BACKUP_AND_RESTORE_ONLY\n' "$0" >&2
  exit 2
fi
if [[ "$CONFIRMATION" != "BACKUP_AND_RESTORE_ONLY" ]]; then
  printf 'Refusing preflight without exact BACKUP_AND_RESTORE_ONLY confirmation.\n' >&2
  exit 2
fi
if [[ ! -f "$ENV_FILE" || -L "$ENV_FILE" ]]; then
  printf 'Refusing preflight: %s must be a regular, non-symlink file.\n' "$ENV_FILE" >&2
  exit 1
fi

env_mode=$(file_mode "$ENV_FILE")
if (( (8#$env_mode & 8#077) != 0 )); then
  printf 'Refusing preflight: %s permissions are %s; require 600 or stricter.\n' "$ENV_FILE" "$env_mode" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

POSTGRES_USER=${POSTGRES_USER:-sub2api}
POSTGRES_DB=${POSTGRES_DB:-sub2api}
TARGET_IMAGE="$IMAGE_REPOSITORY@$IMAGE_DIGEST"
run_id="$(date -u +%Y%m%dT%H%M%SZ)-$$"
backup_path="$BACKUP_DIR/sub2api-$run_id.dump"
report_path="$BACKUP_DIR/sub2api-$run_id-financialgate.json"
restore_container="sub2api-preflight-postgres-$run_id"
restore_network="sub2api-preflight-$run_id"
restore_password=$(openssl rand -hex 24)

cleanup() {
  docker rm -f "$restore_container" >/dev/null 2>&1 || true
  docker network rm "$restore_network" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

for container in "$PROD_APP_CONTAINER" "$PROD_POSTGRES_CONTAINER"; do
  state=$(docker inspect "$container" --format '{{.State.Status}}' 2>/dev/null || true)
  if [[ "$state" != "running" ]]; then
    printf 'Refusing preflight: production container %s is %s.\n' "$container" "${state:-missing}" >&2
    exit 1
  fi
done

server_version=$(docker exec "$PROD_POSTGRES_CONTAINER" \
  psql -XAt -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c 'SHOW server_version')
server_major=${server_version%%.*}
if [[ "$server_major" != "18" ]]; then
  printf 'Refusing preflight: expected PostgreSQL 18, found %s.\n' "$server_version" >&2
  exit 1
fi

db_bytes=$(docker exec "$PROD_POSTGRES_CONTAINER" \
  psql -XAt -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -c "SELECT pg_database_size(current_database())")
user_count=$(docker exec "$PROD_POSTGRES_CONTAINER" \
  psql -XAt -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -c "SELECT CASE WHEN to_regclass('users') IS NULL THEN 0 ELSE (SELECT COUNT(*) FROM users) END")
migration_count=$(docker exec "$PROD_POSTGRES_CONTAINER" \
  psql -XAt -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -c "SELECT CASE WHEN to_regclass('schema_migrations') IS NULL THEN 0 ELSE (SELECT COUNT(*) FROM schema_migrations) END")
available_bytes=$(available_bytes_for "$DEPLOY_DIR")
required_bytes=$((db_bytes * 3 + 1073741824))
if (( available_bytes < required_bytes )); then
  printf 'Refusing preflight: insufficient disk (available=%s required=%s).\n' "$available_bytes" "$required_bytes" >&2
  exit 1
fi

mkdir -p "$BACKUP_DIR"
docker exec "$PROD_POSTGRES_CONTAINER" \
  pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  --format=custom --compress=6 --no-owner --no-acl > "$backup_path"
if [[ ! -s "$backup_path" ]]; then
  printf 'Logical backup is empty: %s\n' "$backup_path" >&2
  exit 1
fi
backup_sha256=$(sha256_file "$backup_path")
printf 'Backup created before candidate validation: %s\n' "$backup_path"
printf 'Backup SHA-256: %s\n' "$backup_sha256"

docker pull "$TARGET_IMAGE" >/dev/null
docker pull "$RESTORE_IMAGE" >/dev/null
docker network create "$restore_network" >/dev/null
docker run -d --name "$restore_container" --network "$restore_network" \
  --label sub2api.preflight=true \
  -e POSTGRES_USER=preflight \
  -e POSTGRES_PASSWORD="$restore_password" \
  -e POSTGRES_DB=sub2api_restore \
  "$RESTORE_IMAGE" >/dev/null

deadline=$((SECONDS + RESTORE_TIMEOUT))
until docker exec "$restore_container" pg_isready -U preflight -d sub2api_restore >/dev/null 2>&1; do
  if (( SECONDS >= deadline )); then
    printf 'Timed out waiting for isolated restore database.\n' >&2
    exit 1
  fi
  sleep 2
done

docker exec -i "$restore_container" pg_restore \
  -U preflight -d sub2api_restore --exit-on-error --single-transaction \
  --no-owner --no-acl < "$backup_path"

restored_user_count=$(docker exec "$restore_container" \
  psql -XAt -U preflight -d sub2api_restore -c 'SELECT COUNT(*) FROM users')
restored_migration_count=$(docker exec "$restore_container" \
  psql -XAt -U preflight -d sub2api_restore -c 'SELECT COUNT(*) FROM schema_migrations')
if [[ "$restored_user_count" != "$user_count" || "$restored_migration_count" != "$migration_count" ]]; then
  printf 'Restore count mismatch: users %s/%s, migrations %s/%s.\n' \
    "$restored_user_count" "$user_count" "$restored_migration_count" "$migration_count" >&2
  exit 1
fi

if ! docker run --rm --network "$restore_network" --entrypoint /app/financialgate \
  -e FINANCIAL_GATE_ALLOW_NON_LOCAL=true \
  "$TARGET_IMAGE" \
  -database-url "postgres://preflight:$restore_password@$restore_container:5432/sub2api_restore?sslmode=disable" \
  -timeout 10m > "$report_path"; then
  printf 'Financial gate failed against the isolated restore. Production was not modified.\n' >&2
  printf 'Credit bucket mismatch details (isolated restore only):\n' >&2
  docker exec "$restore_container" psql -X -U preflight -d sub2api_restore \
    -c "SELECT u.id AS user_id, u.balance AS legacy_balance, a.transferable_credit, a.non_transferable_credit, a.debt, u.balance - (a.transferable_credit + a.non_transferable_credit - a.debt) AS difference FROM users u JOIN user_credit_accounts a ON a.user_id = u.id WHERE u.balance <> a.transferable_credit + a.non_transferable_credit - a.debt ORDER BY u.id LIMIT 20" \
    >&2 || true
  printf 'Backup retained at %s (SHA-256: %s).\n' "$backup_path" "$backup_sha256" >&2
  exit 1
fi

if ! jq -e '.reconciliation | to_entries | all(.value == 0)' "$report_path" >/dev/null; then
  printf 'Financial reconciliation failed; report retained at %s.\n' "$report_path" >&2
  exit 1
fi

printf 'Production source was not migrated or modified.\n'
printf 'Candidate image: %s\n' "$TARGET_IMAGE"
printf 'PostgreSQL: %s; users: %s; migrations before restore gate: %s\n' \
  "$server_version" "$user_count" "$migration_count"
printf 'Backup: %s\n' "$backup_path"
printf 'Backup SHA-256: %s\n' "$backup_sha256"
printf 'Financial gate report: %s\n' "$report_path"
