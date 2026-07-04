---
sidebar_position: 2
title: 'Tutorial: Migrating from Heroku'
---

# Tutorial — Migrating an app from Heroku to BytePort

Most Heroku apps map to BytePort's **process model** with one
declarative manifest. This tutorial walks through a real-world
migration of a SvelteKit + Postgres + Redis app.

**Total time: ~45 minutes.**

> **Prereqs**
> - A working Heroku app (we'll use `myapp-staging` as an example)
> - `heroku-cli` and `byteport` (1.0+) installed
> - Read [First IaC project](first-iac-project.md) if you haven't

## 1. Inventory — capture current Heroku state

```bash
heroku config --shell -a myapp-staging > /tmp/heroku-env.txt
heroku addons --json -a myapp-staging > /tmp/heroku-addons.json
heroku pg:backups:capture -a myapp-staging
```

You'll end up with:
- **env vars** (Doppler/Heroku config → BytePort env registry)
- **addon metadata** (Postgres plans, Redis version, etc.)
- **DB dump** (`.dump` file)

## 2. Translate — Heroku commands → BytePort equivalents

| Heroku                                  | BytePort                                                       |
|-----------------------------------------|----------------------------------------------------------------|
| `heroku config:set X=y`                 | `byteport env set X=y` or `/dashboard/projects/:id/env`        |
| `heroku addons:create heroku-postgresql`| Subscribe via BytePort dashboard → Add-ons tab                 |
| `heroku ps:scale web=2`                 | `byteport.scaling.toml` → `[web] count = 2`                   |
| `heroku logs --tail`                    | `byteport tail` or `/dashboard/deploys/:id/logs`               |
| `heroku releases:rollback`              | `byteport rollback` (one-step, instant)                        |

## 3. Postgres migration — dump-and-stream

```bash
# 1. Capture
heroku pg:backups:download -a myapp-staging
# 2. Restore to BytePort's managed Postgres
pg_restore --no-owner --clean --if-exists \
  -h $BYTEPORT_PG_HOST -U $BYTEPORT_PG_USER \
  -d myapp latest.dump
```

BytePort's managed Postgres is wire-compatible with Heroku's — same
extensions, same `DATABASE_URL` env var semantics.

## 4. Custom domains — route 53 + wildcard certs

```bash
byteport domains add myapp.example.com
# Output:
#   CNAME myapp.example.com -> edge.lax1.byteport.example.com.
#   ACME challenge TXT _acme-challenge.myapp.example.com -> <token>
```

After the DNS TXT records propagate, BytePort automatically provisions
a Let's Encrypt cert and rolls it into the edge. No human action needed.

## 5. Validation — read-only parity check

```bash
byteport parity myapp-staging      # what we call "smoke-test" — runs the
                                   # app's own /health endpoints on both
                                   # platforms and diffs the responses
```

## What you should have observed

- Zero downtime (the cutover is DNS-only)
- All Heroku env vars migrated with one command (Doppler → BytePort env registry uses the same import format)
- Postgres 16 vs Heroku Postgres 16 — same extensions, same wire protocol

## What changes for the developer

| Concern                  | Before (Heroku)                 | After (BytePort)                                        |
|--------------------------|---------------------------------|---------------------------------------------------------|
| Deploy latency           | 60-120s (slug + push)           | 8-15s (incremental snapshot restore)                    |
| Cold start               | 5-15s                           | <100ms (warm pool)                                      |
| Cost visibility          | Post-hoc invoice                | Per-deploy breakdown in `/dashboard/billing`            |
| Custom build plans       | buildpacks only                 | pluggable via `byteport.scaling.toml`                   |

## Next tutorials

- [Custom build plans](custom-build-plans.md) — Rust services, monorepos
- [First IaC project](first-iac-project.md) — single-deployable primer
