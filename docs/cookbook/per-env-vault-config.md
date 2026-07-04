---
sidebar_position: 7
---

# Recipe — Per-environment config with Vault

## What you'll build

Pull secrets from HashiCorp Vault into deploys, with per-environment
isolation and audit log.

**Time: ~18 minutes.**

## Prereqs

- Vault 1.14+ reachable from the BytePort region
- A Vault token (AppRole) with read access to `secret/data/byteport/*`
- A service-account policy granting the `vault:read` scope

## Steps

```bash
# 1. In Vault — create the policy
vault policy write byteport-read - <<'HCL'
path "secret/data/byteport/*" {
  capabilities = ["read"]
}
HCL

# 2. Bind the AppRole
vault write auth/approle/role/byteport \
  token_policies=byteport-read \
  token_ttl=30m \
  token_max_ttl=2h

# 3. In BytePort — register the Vault source
byteport secrets sources add vault \
  --url $VAULT_ADDR \
  --role-id $APPROLE_ROLE_ID \
  --secret-id $APPROLE_SECRET_ID \
  --mount secret \
  --path-prefix byteport

# 4. Reference Vault paths in your project.toml
cat <<'TOML' > project.toml
[secrets]
DATABASE_URL      = "vault://production/database/url"
STRIPE_SECRET_KEY = "vault://production/stripe/key"
SESSION_SECRET    = "vault://production/session/secret"
TOML

# 5. Deploy
byteport deploys create --env production --commit HEAD
```

## Per-environment isolation

Use distinct paths in Vault:
- `secret/data/byteport/production/...`
- `secret/data/byteport/staging/...`
- `secret/data/byteport/dev/...`

BytePort selects the right path based on the deploy's target environment.

## Audit log

Every Vault read is logged in:
- **Vault audit log** (by Vault itself)
- **BytePort deploy log** (by env, project, deploy ID)

View via:
```bash
byteport secrets audit --env production --since 24h
```

## Cleanup

```bash
byteport secrets sources remove vault
vault token revoke $BYTePORT_VAULT_TOKEN
vault auth/approle/role/byteport
```

## Related

- [Custom OAuth provider](custom-oauth-provider.md) — store client secret in Vault
- [Monitoring hook → Grafana](monitoring-hook-grafana.md) — Vault-stored API keys
- Cookbook category: **Security**
