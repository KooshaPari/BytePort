---
sidebar_position: 4
title: Cookbook
---

# Cookbook

Short, focused recipes for specific tasks. Each is a copy-paste-ready
solution with the **why**, the **what to verify**, and links to
related concepts.

## Recipes

### Operations

| # | Recipe | Description |
|---|--------|-------------|
| 1 | [Backup Postgres to S3](postgres-backup-to-s3.md) | Schedule + verify Postgres logical backups to S3 |
| 2 | [Custom domain + wildcard TLS](custom-domain-wildcard-tls.md) | Add `*.example.com` with auto-provisioned cert |
| 3 | [Custom OAuth provider](custom-oauth-provider.md) | Wire an enterprise SSO (Okta/Auth0/Azure AD/Cognito) |
| 4 | [Monitoring hook → Grafana](monitoring-hook-grafana.md) | Emit BytePort metrics to a self-hosted Grafana |
| 5 | [Per-region active-active](per-region-active-active.md) | Multi-region failover with health-checked latency |
| 6 | [Plan-time secrets via Vault](vault-secrets-at-plan-time.md) | Pull build-time secrets (license keys, repo tokens) from Vault |
| 7 | [Cost cap by plan](cost-cap-by-plan.md) | Soft/hard caps per project — kill the next deploy if over budget |
| 8 | [Run a cron](run-a-cron.md) | Time-triggered plans (daily digest, weekly cleanup, hourly poller) |
| 9 | [Blue/green via DNS](blue-green-via-dns.md) | Zero-downtime scheme swaps across the wildcards |
| 10 | [Local dev mirror](local-dev-mirror.md) | Mirror the entire BytePort topology on a laptop with `byteport dev` |

## Categories

- **Compute** — recipes 1, 5, 7
- **Networking** — recipes 2, 5, 9
- **Identity** — recipe 3
- **Observability** — recipes 1, 4
- **Secrets** — recipe 6
- **Economics** — recipe 7
- **Scheduling** — recipe 8
- **Local dev** — recipe 10

## Conventions

Each recipe follows this structure:
1. **What you'll build** (1 sentence)
2. **Prereqs** (what to have ready)
3. **Steps** (numbered, copy-paste-ready)
4. **Verify** (how to confirm it works)
5. **Related** (links to docs/concepts)
6. **Cleanup** (how to undo)
