---
sidebar_position: 10
---

# Recipe — Self-host BytePort on a single VPS

## What you'll build

Run the entire BytePort stack (frontend, backend, scheduler, Postgres,
Redis) on a single $20/month VPS as a private, single-tenant install.

**Time: ~25 minutes.**

## Prereqs

- 1 × VPS (Hetzner CCX23 or equivalent: 4 vCPU, 16 GB RAM, 80 GB NVMe)
- Ubuntu 24.04 LTS
- A DNS A record pointing your domain to the VPS IP
- Outbound SMTP (Mailgun free tier or equivalent)

## Steps

```bash
# 1. Bootstrap
ssh root@$VPS_IP
curl -fsSL byteport.dev/install.sh | sh -s -- \
  --domain apps.example.com \
  --email ops@example.com

# 2. Configure SMTP (interactive prompts)
byteport config smtp
#   host: smtp.mailgun.org
#   user: postmaster@mg.example.com
#   pass: <mailgun-smtp-password>

# 3. Configure TLS (auto, via Caddy)
byteport config tls
#   Domain: apps.example.com
#   DNS:   A apps.example.com → 203.0.113.42  (already pointing)
#   Issuing cert... done.

# 4. Create your admin user
byteport users create admin@example.com --role owner

# 5. Open in browser
#    https://apps.example.com/login
```

## What got installed

| Component | Path | Resource |
|-----------|------|----------|
| Web (Caddy + SvelteKit SSR) | `/opt/byteport/web` | 200 MB RAM |
| API (Go) | `/opt/byteport/api` | 400 MB RAM |
| Scheduler | `/opt/byteport/scheduler` | 150 MB RAM |
| Postgres 16 | `/var/lib/postgresql` | 600 MB RAM |
| Redis 7 | `/var/lib/redis` | 100 MB RAM |
| Caddy (reverse proxy) | `/opt/caddy` | 80 MB RAM |
| Logs / metrics / traces | `/var/log/byteport` + `/var/lib/byteport/observability` | disk |

Total ~1.5 GB RAM, ~3 GB disk + audit log rotation.

## Backups

```bash
# Daily full backup to S3 (built-in)
byteport backup configure s3 \
  --bucket byteport-backups-prod \
  --prefix $(date +%Y/%m/%d) \
  --schedule "0 3 * * *"

# Manual restore
byteport backup restore s3://byteport-backups-prod/2026/07/03/0300
```

## Upgrade

```bash
byteport upgrade check   # shows pending version
byteport upgrade apply   # rolling restart with zero downtime
```

## Monitoring

A built-in Grafana dashboard is available at
`https://apps.example.com/admin/grafana` (owner role required).

## Cleanup

```bash
ssh root@$VPS_IP
byteport uninstall
# (does NOT remove data — add --purge-data to wipe)
```

## Related

- [Postgres backup to S3](postgres-backup-to-s3.md) — backup strategy
- Cookbook category: **Self-hosting**