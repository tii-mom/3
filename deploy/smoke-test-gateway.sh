#!/usr/bin/env bash
# Production smoke tests for api.3api.shop (run on VPS with docker + psql access).
# Does not print full API keys.
set -euo pipefail

BASE_URL="${BASE_URL:-https://api.3api.shop}"
API_KEY_NAME="${API_KEY_NAME:-k12}"
MODEL="${MODEL:-gpt-5.6-sol}"
PG_CONTAINER="${PG_CONTAINER:-sub2api-postgres}"
PG_USER="${PG_USER:-sub2api}"
PG_DB="${PG_DB:-sub2api}"

pass=0
fail=0

check() {
  local name="$1" cond="$2"
  if eval "$cond"; then
    echo "PASS  $name"
    pass=$((pass + 1))
  else
    echo "FAIL  $name"
    fail=$((fail + 1))
  fi
}

echo "== health =="
code=$(curl -sS -o /tmp/sm_health.json -w '%{http_code}' --max-time 15 "$BASE_URL/health" || echo 000)
check "health 200" "[[ \"$code\" == \"200\" ]]"
check "health body ok" "grep -q '\"status\":\"ok\"' /tmp/sm_health.json 2>/dev/null"

echo "== nginx body limit =="
if [[ -f /etc/nginx/conf.d/sub2api.conf ]]; then
  check "client_max_body_size set" "grep -qE 'client_max_body_size\\s+[0-9]+m' /etc/nginx/conf.d/sub2api.conf"
else
  echo "SKIP  nginx conf not on this host"
fi

echo "== account pool SQL =="
if docker ps --format '{{.Names}}' | grep -qx "$PG_CONTAINER"; then
  docker exec "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -c "
SELECT g.id, g.name, g.platform,
       COUNT(a.id) FILTER (WHERE a.deleted_at IS NULL) AS linked,
       COUNT(a.id) FILTER (WHERE a.deleted_at IS NULL AND a.status='active' AND a.schedulable) AS schedulable
FROM groups g
LEFT JOIN account_groups ag ON ag.group_id = g.id
LEFT JOIN accounts a ON a.id = ag.account_id
WHERE g.deleted_at IS NULL
GROUP BY g.id
ORDER BY g.id;"
  docker exec "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -c "
SELECT id, name, status, schedulable,
       (credentials ? 'refresh_token') AS has_rt,
       (proxy_id IS NOT NULL) AS has_proxy
FROM accounts WHERE deleted_at IS NULL;"
else
  echo "SKIP  postgres container not found"
fi

KEY=""
if docker ps --format '{{.Names}}' | grep -qx "$PG_CONTAINER"; then
  KEY=$(docker exec "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -tAc \
    "SELECT key FROM api_keys WHERE name='${API_KEY_NAME}' AND deleted_at IS NULL LIMIT 1" | tr -d '[:space:]')
fi

if [[ -z "$KEY" ]]; then
  echo "SKIP  API key name='$API_KEY_NAME' not found — set API_KEY_NAME or insert key"
else
  echo "== short /responses (key name=$API_KEY_NAME) =="
  code=$(curl -sS -o /tmp/sm_resp.json -w '%{http_code}' --max-time 90 \
    -X POST "$BASE_URL/responses" \
    -H "Authorization: Bearer $KEY" \
    -H 'Content-Type: application/json' \
    -d "{\"model\":\"$MODEL\",\"input\":\"ping smoke\",\"stream\":false}" || echo 000)
  check "responses HTTP 200" "[[ \"$code\" == \"200\" ]]"
  if [[ "$code" != "200" ]]; then
    head -c 300 /tmp/sm_resp.json; echo
  fi

  echo "== 1.5MB body (must not be nginx 413 HTML) =="
  python3 - <<'PY'
import json
pad = "z" * (1500 * 1024)
open("/tmp/sm_big.json", "w").write(json.dumps({
    "model": __import__("os").environ.get("MODEL", "gpt-5.6-sol"),
    "input": "smoke " + pad,
    "stream": False,
    "max_output_tokens": 8,
}))
print("size", __import__("os").path.getsize("/tmp/sm_big.json"))
PY
  MODEL="$MODEL" code=$(curl -sS -o /tmp/sm_big.out -w '%{http_code}' --max-time 90 \
    -X POST "$BASE_URL/responses" \
    -H "Authorization: Bearer $KEY" \
    -H 'Content-Type: application/json' \
    --data-binary @/tmp/sm_big.json || echo 000)
  is_nginx_413=0
  if grep -qi 'Request Entity Too Large' /tmp/sm_big.out 2>/dev/null; then
    is_nginx_413=1
  fi
  check "1.5MB not nginx-413" "[[ \"$is_nginx_413\" -eq 0 ]]"
  echo "  (HTTP $code — 200 ok; 502 context-window is upstream, not nginx 413)"
fi

echo
echo "Summary: pass=$pass fail=$fail"
[[ "$fail" -eq 0 ]]
