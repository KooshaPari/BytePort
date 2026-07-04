---
sidebar_position: 3
---

# Recipe — Custom OAuth provider

## What you'll build

Wire an enterprise SSO provider (Okta / Auth0 / Azure AD / Cognito)
as a login path on your BytePort workspace.

**Time: ~15 minutes.**

## Prereqs

- Owner access to your workspace
- A test user in the SSO provider
- The SSO provider's discovery URL (`https://<tenant>/.well-known/openid-configuration`)

## Steps

```bash
# 1. Register the OIDC app in your SSO provider
#    Redirect URI: https://byteport.example.com/api/v1/auth/oauth/<provider>/callback
#    Grant type:   authorization_code
#    Scopes:       openid profile email groups

# 2. Add the provider config to BytePort
byteport oauth providers add okta \
  --workspace $WORKSPACE \
  --issuer 'https://dev-12345.okta.com/oauth2/default' \
  --client-id '0oab12345XYZ' \
  --client-secret "$(byteport vault read 'vault:kv/oauth/okta/client-secret')" \
  --default-roles '{"engineer": "developer", "viewer": "viewer"}'

# 3. Map Okta groups → BytePort roles (one-time)
cat <<'TOML' >> .byteport/role-map.toml
[oauth.okta.role-map]
"Engineers"            = "developer"
"SRE"                  = "admin"
"Engineering-Managers" = "lead"
"Everyone"             = "viewer"
TOML
byteport oauth providers validate --name okta

# 4. Publish (gates users without `Engineering-*` group out of any sensitive deploy)
byteport oauth providers publish --name okta --workspace $WORKSPACE
```

## Verify

```bash
# 1. Browser flow
open "https://byteport.example.com/login?provider=okta"

# 2. CLI flow (for users without browser)
byteport login --provider okta --refresh   # opens browser then closes
```

## SAML alternative

For SAML 2.0 IdPs (Okta, Auth0, OneLogin, etc.):

```bash
byteport oauth providers add okta-saml \
  --workspace $WORKSPACE \
  --saml-metadata-url 'https://dev-12345.okta.com/app/<app-id>/sso/saml/metadata' \
  --default-roles '{"Engineers": "developer"}'
```

## Cleanup

```bash
byteport oauth providers remove okta --workspace $WORKSPACE
```

## Related

- [Vault secrets at plan time](vault-secrets-at-plan-time.md) — store provider secrets in Vault
- Cookbook category: **Identity**
