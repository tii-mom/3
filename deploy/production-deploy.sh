#!/usr/bin/env bash

set -Eeuo pipefail

DEPLOY_DIR=${DEPLOY_DIR:-/opt/sub2api-deploy}
COMPOSE_FILE=${COMPOSE_FILE:-$DEPLOY_DIR/docker-compose.local.yml}
IMAGE_REPOSITORY=${IMAGE_REPOSITORY:-ghcr.io/tii-mom/3}
IMAGE_TAG=${1:-latest}
TARGET_IMAGE="$IMAGE_REPOSITORY:$IMAGE_TAG"
HEALTH_TIMEOUT=${HEALTH_TIMEOUT:-120}

cd "$DEPLOY_DIR"

validate_env_permissions() {
  if [[ ! -f .env || -L .env ]]; then
    printf 'Refusing deployment: %s/.env must be a regular, non-symlink file.\n' "$DEPLOY_DIR" >&2
    exit 1
  fi

  local mode owner current_uid
  mode=$(stat -c '%a' .env)
  owner=$(stat -c '%u' .env)
  current_uid=$(id -u)
  if (( (8#$mode & 8#077) != 0 )); then
    printf 'Refusing deployment: .env permissions are %s; require 600 or stricter.\n' "$mode" >&2
    exit 1
  fi
  if [[ "$current_uid" != "0" && "$owner" != "$current_uid" ]]; then
    printf 'Refusing deployment: .env must be owned by the deployment user.\n' >&2
    exit 1
  fi
}

validate_env_permissions

persist_image() {
  local image=$1
  if grep -q '^SUB2API_IMAGE=' .env; then
    sed -i "s|^SUB2API_IMAGE=.*|SUB2API_IMAGE=$image|" .env
  else
    printf '\nSUB2API_IMAGE=%s\n' "$image" >> .env
  fi
}

previous_image=$(docker inspect sub2api --format '{{.Config.Image}}' 2>/dev/null || true)
export SUB2API_IMAGE="$TARGET_IMAGE"

docker pull "$TARGET_IMAGE"
docker compose -f "$COMPOSE_FILE" up -d --no-deps --force-recreate sub2api

deadline=$((SECONDS + HEALTH_TIMEOUT))
while (( SECONDS < deadline )); do
  status=$(docker inspect sub2api --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' 2>/dev/null || true)
  if [[ "$status" == "healthy" ]]; then
    persist_image "$TARGET_IMAGE"
    printf 'Deployed %s successfully.\n' "$TARGET_IMAGE"
    exit 0
  fi
  if [[ "$status" == "unhealthy" || "$status" == "exited" || "$status" == "dead" ]]; then
    break
  fi
  sleep 3
done

printf 'Deployment failed health checks for %s.\n' "$TARGET_IMAGE" >&2
docker logs --tail 100 sub2api >&2 || true

if [[ -n "$previous_image" ]]; then
  printf 'Rolling back to %s.\n' "$previous_image" >&2
  persist_image "$previous_image"
  export SUB2API_IMAGE="$previous_image"
  docker compose -f "$COMPOSE_FILE" up -d --no-deps --force-recreate sub2api
fi

exit 1
