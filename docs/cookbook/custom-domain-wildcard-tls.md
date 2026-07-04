---
sidebar_position: 2
---

# Recipe — Custom domain + wildcard TLS

## What you'll build

Bind `*.example.com` (and optionally `example.com`) to your
BytePort project and have a **Let's Encrypt** wildcard cert
provisioned automatically.

**Time: ~8 minutes.**

## Prereqs

- A BytePort project with at least one successful deploy
- `octo` or `byteport` CLI 1.0+
- A registered domain in your DNS provider
- Cloudflare account (free tier is fine) — used for the ACME DNS-01 challenge

## Steps

```bash
# 1. Add the domain to the project
byteport domains add '*.example.com' --project :id --challenge dns-01
# Output:
#   Two CNAME records + one TXT record to add at your DNS provider:
#     _acme-challenge.example.com.   TXT   "<token>"
#     _acme-challenge.example.com.   CNAME _acme-challenge.acme.byteport.example.com.
#     example.com. (wildcard)        NS    ns1.byteport.example.com.

# 2. Add the records at Cloudflare (DNS app → Add record)
# Verify with:
byteport domains verify '*.example.com' --project :id

# 3. Wait for NS propagation (10 min - 24h depending on registrar)
# BytePort's CLI shows progress:
byteport domains watch '*.example.com'

# 4. Once propagated, BytePort auto-provisions the wildcard cert
# Visible in /dashboard/projects/:id/domains
```

## Verify

```bash
curl -I https://anything.example.com       # should be 200/404 from your app, NOT 523 from BytePort
openssl s_client -connect anything.example.com:443 -servername example.com \
  | openssl x509 -text -noout | grep -A2 'Subject Alternative Name'
# Should show: DNS:example.com, DNS:*.example.com
```

## How long does it take?

- DNS records added at Cloudflare: instant
- ACME challenge (TXT validation): 1-2 minutes
- Wildcard cert issuance from Let's Encrypt: another 1-2 minutes
- NS delegation + visible `*.example.com`: 5 minutes - 24h (registrar TTL)

## Cleanup

```bash
byteport domains remove '*.example.com' --project :id
# Note: cert remains valid until expiry (~80 days), then auto-revoked
```

## Related

- [Per-region active-active](per-region-active-active.md) — multi-region wildcard
- [Blue/green via DNS](blue-green-via-dns.md) — scheme swap with the same wildcard
- Cookbook category: **Networking**
