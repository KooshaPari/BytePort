---
sidebar_position: 8
---

# Recipe — Scheduled jobs (cron) inside a project

## What you'll build

Run a periodic task inside your project (e.g. "send digest emails at
8am UTC every day", "rebuild search index every 15 min") without
needing a separate worker.

**Time: ~10 minutes.**

## Steps

```bash
# 1. Define the schedule
cat <<'TOML' >> project.toml
[[schedules]]
name    = "send-digest"
command = "rails runner 'DigestMailer.daily'"
cron    = "0 8 * * *"
tz      = "UTC"
retries = 3

[[schedules]]
name    = "reindex"
command = "node scripts/reindex.js"
every   = "15m"
retries = 1
```

## Schedule formats

| Field | Example | Meaning |
|-------|---------|---------|
| `cron` | `0 8 * * *` | standard 5-field cron (minute hour dom month dow) |
| `every` | `15m` / `2h` / `1d` | duration interval |
| `tz` | `UTC` / `America/Los_Angeles` | timezone for cron schedules |
| `at` | `2026-08-01T00:00:00Z` | one-shot (must be future) |

## Lifecycle

1. **Scaled to zero** — no cost between runs
2. **Cold-start budget** — first run after idle may add ~1s for container spin-up
3. **Concurrency guard** — if a previous run is still in flight, the new run is skipped (not stacked)
4. **Retries** — N attempts with exponential backoff (max 5m between attempts)
5. **Failure alert** — after retries exhausted, fires `schedule.failure` webhook

## View history

```bash
byteport schedules history send-digest --limit 20
```

## Cleanup

```bash
# Disable without removing the definition
byteport schedules disable send-digest

# Remove entirely
# (remove the [[schedules]] block from project.toml and redeploy)
```

## Related

- [Webhook → Discord / Slack](webhook-discord-slack.md) — alert on schedule.failure
- Cookbook category: **Background work**
