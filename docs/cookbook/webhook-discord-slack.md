---
sidebar_position: 5
---

# Recipe — Webhook → Discord / Slack on every deploy

## What you'll build

Post a message to a Discord channel (or Slack channel) every time a
deploy succeeds or fails.

**Time: ~8 minutes.**

## Steps

### Discord

```bash
# 1. Get a webhook URL
# Discord channel → Edit Channel → Integrations → Webhooks → New Webhook
# Copy the URL. Format: https://discord.com/api/webhooks/123/abc...

# 2. Register it in BytePort
byteport hooks add discord-deploys \
  --url "$DISCORD_WEBHOOK_URL" \
  --events deploy.success,deploy.failure \
  --template discord
```

The `discord` template produces a card with embed color (green/red),
title, deployer, commit SHA, and a "View in BytePort" link.

### Slack

```bash
# 1. Get a Slack webhook
# https://api.slack.com/messaging/webhooks → Create New Webhook

# 2. Register
byteport hooks add slack-deploys \
  --url "$SLACK_WEBHOOK_URL" \
  --events deploy.success,deploy.failure \
  --template slack
```

The `slack` template produces a Block-Kit card with a status icon
(✅ / ❌), deployer, commit, and an Actions block with a "View" button.

## Filter events

```bash
# Only fire for production deploys from the main branch
byteport hooks update discord-deploys \
  --when 'env == "production" && ref == "refs/heads/main"'
```

## Verify

Trigger a deploy to a test environment:
```bash
byteport deploys create \
  --project blog \
  --commit HEAD \
  --env staging
```

You should see a message in your Discord/Slack channel within ~3s of the deploy completing.

## Cleanup

```bash
byteport hooks remove discord-deploys
byteport hooks remove slack-deploys
```

## Related

- Cookbook category: **Notifications**
