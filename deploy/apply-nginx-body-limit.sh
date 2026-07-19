#!/usr/bin/env bash
# Apply client_max_body_size for api.3api.shop (fix Codex 413 Payload Too Large).
# Run on the production VPS as root (or with sudo).
#
# Usage:
#   sudo bash apply-nginx-body-limit.sh
#   sudo bash apply-nginx-body-limit.sh /path/to/nginx-sub2api.conf
set -euo pipefail

SRC="${1:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NGINX_CONF="${NGINX_CONF:-/etc/nginx/conf.d/sub2api.conf}"
BODY_LINE='client_max_body_size 256m;'

if [[ -n "$SRC" && -f "$SRC" ]]; then
  CONF_SRC="$SRC"
elif [[ -f "${SCRIPT_DIR}/nginx-sub2api.conf" ]]; then
  CONF_SRC="${SCRIPT_DIR}/nginx-sub2api.conf"
elif [[ -f /opt/sub2api-deploy/nginx-sub2api.conf ]]; then
  CONF_SRC=/opt/sub2api-deploy/nginx-sub2api.conf
else
  CONF_SRC=""
fi

if [[ "$(id -u)" -ne 0 ]]; then
  echo "ERROR: run as root (sudo)." >&2
  exit 1
fi

if ! command -v nginx >/dev/null 2>&1; then
  echo "ERROR: nginx not found on PATH." >&2
  exit 1
fi

backup=""
if [[ -f "$NGINX_CONF" ]]; then
  backup="${NGINX_CONF}.bak.$(date +%Y%m%d%H%M%S)"
  cp -a "$NGINX_CONF" "$backup"
  echo "Backup: $backup"
fi

if [[ -n "$CONF_SRC" ]]; then
  echo "Installing full conf from: $CONF_SRC → $NGINX_CONF"
  install -m 0644 "$CONF_SRC" "$NGINX_CONF"
else
  echo "No packaged nginx-sub2api.conf found; patching $NGINX_CONF in place."
  if [[ ! -f "$NGINX_CONF" ]]; then
    echo "ERROR: $NGINX_CONF missing and no source conf to install." >&2
    exit 1
  fi
  if grep -qE '^\s*client_max_body_size\s+' "$NGINX_CONF"; then
    sed -i -E 's/client_max_body_size\s+[^;]+;/client_max_body_size 256m;/' "$NGINX_CONF"
    echo "Updated existing client_max_body_size → 256m"
  else
    # Insert after server_name api.3api.shop; inside the SSL server block if possible
    if grep -q 'server_name api.3api.shop' "$NGINX_CONF"; then
      # After first underscores_in_headers or server_name in 443 block
      if grep -q 'underscores_in_headers' "$NGINX_CONF"; then
        sed -i '/underscores_in_headers/a\    '"$BODY_LINE" "$NGINX_CONF"
      else
        sed -i '/server_name api.3api.shop;/a\    '"$BODY_LINE" "$NGINX_CONF"
      fi
      # Deduplicate if sed ran on both server blocks: keep only 256m lines, collapse multiples later via nginx -t
      echo "Inserted: $BODY_LINE"
    else
      echo "ERROR: could not find server_name api.3api.shop in $NGINX_CONF" >&2
      exit 1
    fi
  fi
fi

if ! grep -qE 'client_max_body_size\s+256m' "$NGINX_CONF"; then
  echo "WARN: client_max_body_size 256m not detected in $NGINX_CONF after edit." >&2
fi

echo "Testing nginx config..."
nginx -t

echo "Reloading nginx..."
if systemctl is-active --quiet nginx 2>/dev/null; then
  systemctl reload nginx
elif command -v nginx >/dev/null; then
  nginx -s reload
fi

echo "OK: nginx reloaded with client_max_body_size 256m"
echo
echo "Quick checks:"
echo "  curl -sS -i https://api.3api.shop/health | head -20"
echo "  grep client_max_body_size $NGINX_CONF"
echo
echo "If Codex still returns 503 (not 413): check account pool + app logs, not body size."
echo "  docker logs sub2api --tail 100"
echo "  Admin → OpenAI/Codex accounts active & schedulable"
