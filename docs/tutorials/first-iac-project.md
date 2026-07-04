---
sidebar_position: 1
title: 'Tutorial: First IaC project'
---

# Tutorial — Your First IaC project on BytePort

This tutorial walks you through the **single deployable unit** BytePort is
optimized for: a SvelteKit site that builds on push, deploys to a fresh
microVM, and gets a wildcard TLS certificate automatically.

**Total time: ~30 minutes.**

> **Prereqs**
> - BytePort instance running (`brew install byteport && byteport server start`)
> - A git repo with a SvelteKit project (you'll have one once you `pnpm create svelte@latest`)
> - This tutorial uses `byteport` CLI 1.0+, available via the "CLI" tab in `/dashboard/cli`

## 1. Initialize — bind a project to a project

```bash
cd my-svelte-site
byteport init
```

This writes `.byteport/config.toml` with:
- **Project name** (auto-detected from `package.json` → `name`)
- **Build plan** (auto-detected from SvelteKit adapter — `adapter-auto` is supported out of the box)
- **Default region** (use the prompt — `lax1` or `iad1` are usually fastest)

## 2. Connect — link the project to your BytePort workspace

```bash
byteport login        # one-time — opens browser for SSO
byteport connect      # prints the project ID + a webhook URL
```

Copy the webhook URL into your git remote:

```bash
git remote add byteport  https://byteport.example.com/api/v1/projects/<id>/git
git push byteport main  # this triggers your first build
```

## 3. Observe — watch the build progress

```bash
byteport tail         # live logs
```

Or open the `Deploys` tab in the dashboard — pull-to-refresh on mobile, auto-
refresh on desktop.

## 4. Rollback — that's it, one command

```bash
byteport rollback     # restore the previous successful deploy
```

## What you should have observed

- A subdomain like `my-svelte-site.lax1.byteport.example.com`
- Let's Encrypt cert provisioned automatically
- A clean deploy log showing: build → snapshot → boot → health-check → live

## Next tutorials

- [Migrating from Heroku](heroku-migration.md) — multi-process apps, custom domains, observability hooks
- [Custom build plans](custom-build-plans.md) — Rust services, monorepos, multi-stage builds
