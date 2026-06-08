#!/usr/bin/env bash
# Smoke test for the Keycloak auth flow against the running Docker Compose stack.
#
# Usage:
#   docker compose up -d --build      # keycloak imports realm-edu.json, plus db/minio/app
#   ./scripts/smoke.sh
#
# Verifies that Keycloak is the source of truth for identity/roles and that the
# backend trusts it end-to-end:
#   1. Keycloak issues real tokens (Direct Access Grant) for the seeded users
#   2. the backend validates those tokens (JWKS) and serves a protected route
#   3. role from the token is enforced (student -> admin route = 403)
#
# NB: the group_admin's group comes from the Keycloak `group_id` claim
# (mapper in realm-edu.json), i.e. Keycloak — not our DB — owns the binding.
set -euo pipefail

KC="${KC:-http://localhost:8180}"
API="${API:-http://localhost:8080}"
REALM="${REALM:-edu}"
CLIENT="${CLIENT:-edu-frontend}"

command -v jq >/dev/null || { echo "jq is required"; exit 1; }

pass() { printf '  \033[32m✓\033[0m %s\n' "$1"; }
fail() { printf '  \033[31m✗\033[0m %s\n' "$1"; exit 1; }

token() { # username -> access_token
  curl -sf -d "client_id=$CLIENT" -d "username=$1" -d "password=password" \
    -d "grant_type=password" \
    "$KC/realms/$REALM/protocol/openid-connect/token" | jq -r .access_token
}

code() { # METHOD PATH TOKEN -> http status code (body in /tmp/smoke_body)
  curl -s -o /tmp/smoke_body -w '%{http_code}' -X "$1" "$API$2" \
    -H "Authorization: Bearer $3"
}

echo "==> Keycloak issues tokens"
ADMIN=$(token admin1)     && [[ -n "$ADMIN"   && "$ADMIN"   != null ]] && pass "admin1 token"   || fail "admin1 token"
STUDENT=$(token student1) && [[ -n "$STUDENT" && "$STUDENT" != null ]] && pass "student1 token" || fail "student1 token"

echo "==> group_id claim is present in admin1's token (Keycloak owns the binding)"
GID=$(echo "$ADMIN" | cut -d. -f2 | tr '_-' '/+' | { read p; printf '%s' "$p$(printf '=%.0s' $(seq $(( (4 - ${#p} % 4) % 4 ))))"; } | base64 -d 2>/dev/null | jq -r .group_id)
[[ "$GID" == "1" ]] && pass "group_id=1 in token" || fail "expected group_id=1, got '$GID'"

echo "==> backend validates the token and serves a protected route"
c=$(code GET /me/balance "$STUDENT")
[[ "$c" == 200 ]] && pass "student /me/balance = 200 ($(jq -c . /tmp/smoke_body))" || fail "/me/balance = $c"

echo "==> request without a token is rejected"
c=$(curl -s -o /dev/null -w '%{http_code}' "$API/me/balance")
[[ "$c" == 401 ]] && pass "no token -> 401" || fail "expected 401, got $c"

echo "==> role from token is enforced (student -> admin route)"
c=$(code GET /admin/activities "$STUDENT")
[[ "$c" == 403 ]] && pass "student -> /admin/activities = 403" || fail "expected 403, got $c"

echo "==> admin role passes the role gate"
c=$(code GET /admin/activities "$ADMIN")
[[ "$c" == 200 ]] && pass "admin -> /admin/activities = 200" || fail "expected 200, got $c"

printf '\n\033[32mKeycloak auth flow OK.\033[0m\n'
echo "Note: group/course membership in our DB (migrations) still drives the admin feed contents;"
echo "      this script only proves the Keycloak-owned identity/role/group_id flow."
