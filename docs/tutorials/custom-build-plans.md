---
sidebar_position: 3
title: 'Tutorial: Custom build plans'
---

# Tutorial — Custom build plans for monorepos + Rust services

For most apps, BytePort's auto-detected plan is enough. This tutorial
covers **monorepos**, **Rust services**, and **multi-stage builds**.

**Total time: ~20 minutes.**

## 1. The `.byteport/plans.toml` file

If your repo is more complex than a single SvelteKit/Next.js app,
drop a `.byteport/plans.toml` at the root:

```toml
# Build a SvelteKit site AND a Rust worker from one monorepo
[plans.web]
src       = "apps/web"
framework = "sveltekit"
adapters  = ["node"]

[plans.worker]
src         = "services/worker"
framework   = "rust"
binary_name = "byte-worker"
runtime     = "static-rust"
```

BytePort reads this on `git push` and creates **two independent deployables**.
Each gets its own subdomain, its own scaling plan, and its own rollback history.

## 2. Static-binary Rust services

For Rust services that produce a static binary:

```toml
[plans.echo]
src         = "examples/echo"
framework   = "rust"
runtime     = "static-rust"
env         = { RUST_LOG = "info" }
memory      = "64Mi"
count       = 1
```

`static-rust` is a **scratch-style rootfs**: just enough to run your
binary. A 4MiB image. Cold start: <10ms.

## 3. Multi-stage builds with secret injection

```toml
[plans.api]
src                = "apps/api"
framework          = "go"
secret_injection   = true         # secrets injected at BOOT, not build

[plans.api.secrets]
DATABASE_URL       = "vault:kv/api/db"
JWT_SIGNING_KEY    = "vault:kv/api/jwt"
```

`secret_injection = true` means:
- Build logs **never** see the secret values
- The microVM is **fresh-destroyed** on every boot — no disk persistence
- Vault short-lived tokens auto-rotate per boot

## 4. Visual confirmation in the dashboard

- Each plan gets a separate card on `/m/projects/:id/plans`
- The "Build matrix" panel shows you plan × region → which build succeeded
- Click any cell to see the full build log

## Next tutorials

- [First IaC project](first-iac-project.md)
- [Migrating from Heroku](heroku-migration.md)
