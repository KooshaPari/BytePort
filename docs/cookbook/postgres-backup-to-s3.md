---
sidebar_position: 1
---

# Recipe — Backup Postgres to S3

## What you'll build

Schedule a daily logical backup of a BytePort-managed Postgres
instance, push to an S3 bucket, retain for 14 days, and have the
backup result visible in the dashboard.

**Time: ~10 minutes.**

## Prereqs

- A `byteport` CLI 1.0+ authenticated to your workspace
- An S3 bucket you can write to (we'll use `bp-pg-backups-$WORKSPACE`)
- An IAM access key (we'll store it as a Vault token, not env var)

## Steps

```bash
# 1. Create the bucket (one-time)
aws s3api create-bucket --bucket bp-pg-backups-$WORKSPACE --region us-east-1

# 2. Add the IAM access key to Vault (not env vars — see plan-time secrets)
byteport vault set s3-access-key 'vault:kv/s3-access-key'

# 3. Subscribe the add-on for the daily hook
byteport addons subscribe scheduler@1 --plan hourly-10 --project :id
byteport addons subscribe backups@1 --project :id

# 4. Wire the schedule
cat <<'TOML' > .byteport/hooks/postgres-backup.toml
[hook.postgres-backup]
trigger = "schedule:0 3 * * *"            # 3am UTC daily
runtime  = "alpine-3.20"                   # tiny scratch rootfs
command  = ["pg_dump", "$PGDATABASE_URL", "-Fc"]
output   = "s3://bp-pg-backups-$WORKSPACE/daily-$(date +%Y%m%d).dump"
retention_days = 14
TOML
byteport hooks apply

# 5. Test (forces a one-shot run)
byteport hooks run --now postgres-backup

# 6. Verify
byteport hooks list --last postgres-backup   # shows status: success + bytes
```

## Verify

```bash
aws s3 ls s3://bp-pg-backups-$WORKSPACE/    # should show today's .dump
byteport hooks list --project :id          # should show success for today's slot
```

BytePort surfaces a backup **indicator** on the project card:
`✓ last backup 13h ago (4.2 GB)`.

## Cleanup

```bash
byteport hooks remove postgres-backup
aws s3 rm s3://bp-pg-backups-$WORKSPACE/ --recursive
```

## Related

- [Run a cron](run-a-cron.md) — same scheduler for arbitrary commands
- [Vault secrets at plan time](vault-secrets-at-plan-time.md) — how `vault:kv/...` references resolve
- Cookbook category: **Compute**
