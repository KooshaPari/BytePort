---
sidebar_position: 6
---

# Recipe — Staging-clone with scrubbed secrets

## What you'll build

Clone a production database into staging with all PII and secrets
replaced with deterministic fakes.

**Time: ~15 minutes.**

## Prereqs

- A production Postgres database
- A staging environment
- A `scrub` config (replacement rules)

## Steps

```bash
# 1. Define scrub rules
cat <<'YAML' > .byteport/scrub.yaml
tables:
  users:
    email: faker.email      # user-{n}@example.com
    phone: faker.phone
    name:  faker.name
  payments:
    card_number: null       # set NULL
    cvc:        null
sessions:
  hash_secret: random       # rotated after clone
audit_log: drop             # GDPR right-to-erasure
YAML

# 2. Run the clone
byteport db clone \
  --from production \
  --to staging \
  --scrub .byteport/scrub.yaml \
  --snapshot-type consistent

# 3. Confirm
byteport db status --env staging
# age:        3.2s
# row count:  users=24_103, payments=58_204 (matches prod)
# pii audit:  passed (0 unscrubbed columns)
```

## What gets scrubbed

| Type | Strategy |
|------|----------|
| Email | `user-{n}@example.com` (deterministic) |
| Phone | `+1555{n:07d}` (reserved test range) |
| Name | Faker from `pg-faker` |
| Free text | Reversible MD5 hash with tenant salt |
| Secret columns | `NULL` (must be explicitly listed) |
| Audit log | `DROP TABLE` post-import |

## Why deterministic?

Deterministic scrubbing means foreign keys still work. `user.id=42`'s
`email` becomes `user-42@example.com`, so all joins in your staging
tests are reproducible.

## Cleanup

The clone runs entirely in the staging environment — nothing to clean
up in production.

## Related

- [Postgres backup to S3](postgres-backup-to-s3.md) — for the source snapshot
- [Custom OAuth provider](custom-oauth-provider.md) — test auth in staging with mock provider
- Cookbook category: **Data**
