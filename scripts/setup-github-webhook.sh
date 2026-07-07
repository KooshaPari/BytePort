#!/usr/bin/env bash
#
# setup-github-webhook.sh — Register BytePort's GitHub push webhook on a repo.
#
# This wires the PoC-D webhook handler (backend/internal/infrastructure/http/handlers/webhook_handler.go)
# into real CI: on every `push` to $BRANCH, GitHub POSTs to
# $BYTEPORT_URL/api/v1/webhooks/github with an HMAC-SHA256 signature that the
# handler verifies against GITHUB_WEBHOOK_SECRET.
#
# Usage:
#   GITHUB_TOKEN=ghp_xxx ./scripts/setup-github-webhook.sh owner/repo
#
# Env (all optional unless noted):
#   GITHUB_TOKEN        (req) token with `admin:repo_hook` on the target repo
#   BYTEPORT_URL        (req) public base URL of the running BytePort API, e.g. https://bp.example.com
#   GITHUB_WEBHOOK_SECRET (req) shared secret; set on the server as GITHUB_WEBHOOK_SECRET too
#   BRANCH              default: main   (only pushes to this branch auto-deploy)
#   ALLOW_UNSIGNED      default: unset  (set to "1" only for local/dev to skip HMAC)
#
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: GITHUB_TOKEN=... BYTEPORT_URL=... $0 owner/repo" >&2
  exit 2
fi

REPO="$1"
BRANCH="${BRANCH:-main}"
BYTEPORT_URL="${BYTEPORT_URL:?set BYTEPORT_URL to your public API base, e.g. https://bp.example.com}"
SECRET="${GITHUB_WEBHOOK_SECRET:?set GITHUB_WEBHOOK_SECRET to the same value the server uses}"
TOKEN="${GITHUB_TOKEN:?set GITHUB_TOKEN with admin:repo_hook scope}"

# Idempotent: delete any existing BytePort webhook on this repo first.
echo ">> scanning existing webhooks on $REPO"
EXISTING=$(curl -sS -f -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "https://api.github.com/repos/$REPO/hooks")

echo "$EXISTING" | python3 -c "
import json, sys
hooks = json.load(sys.stdin)
for h in hooks:
    if h.get('config', {}).get('url', '').endswith('/api/v1/webhooks/github'):
        print(h['id'])
" | while read -r hid; do
  echo ">> removing existing webhook $hid"
  curl -sS -f -X DELETE -H "Authorization: Bearer $TOKEN" \
    -H "Accept: application/vnd.github+json" \
    "https://api.github.com/repos/$REPO/hooks/$hid"
done

echo ">> creating webhook on $REPO -> $BYTEPORT_URL/api/v1/webhooks/github"
curl -sS -f -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/vnd.github+json" \
  -H "Content-Type: application/json" \
  -d "$(python3 -c "
import json, os, sys
print(json.dumps({
  'name': 'web',
  'active': True,
  'events': ['push'],
  'config': {
    'url': os.environ['BYTEPORT_URL'].rstrip('/') + '/api/v1/webhooks/github',
    'content_type': 'json',
    'secret': os.environ['GITHUB_WEBHOOK_SECRET'],
    'insecure_ssl': '0',
  },
}))")" \
  "https://api.github.com/repos/$REPO/hooks"

echo ">> done. Pushes to '$BRANCH' on $REPO will now trigger BytePort deployments."
echo "   Server must run with: GITHUB_WEBHOOK_SECRET=$SECRET GITHUB_WEBHOOK_BRANCH=$BRANCH"
