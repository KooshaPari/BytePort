---
sidebar_position: 9
---

# Recipe — Local dev → BytePort preview in 30 seconds

## What you'll build

A one-command workflow that pushes your local dev branch to an
ephemeral preview environment you can share with a teammate.

**Time: ~2 minutes** (after first-time setup).

## Setup (once)

```bash
# Install the BytePort CLI
brew install byteport
# or: scoop install byteport
# or: curl -L byteport.dev/install.sh | sh

# Login
byteport auth login

# Pin your project
byteport project use my-app
```

## Workflow

```bash
# 1. Create a branch
git checkout -b feat/new-thing

# 2. ...make changes, commit...

# 3. Ship a preview (uses your branch name)
byteport preview push
#   → preview-my-app-feat-new-thing.byteport.app (HTTPS, shareable, auto-destroyed in 7 days)

# 4. Iterate — each push updates the same preview
git commit -m "more changes"
byteport preview push

# 5. Clean up when done
byteport preview remove
```

## What it does

| Step | What happens |
|------|--------------|
| `git push` (local → remote) | unchanged, your normal workflow |
| `byteport preview push` | triggers CI on a new ephemeral env |
| env = preview | copies the `staging` config + replaces secrets with ephemeral test secrets |
| DNS | wildcard `*.byteport.app` automatically routes to the preview container |
| TLS | auto-issued via Let's Encrypt, no human action |
| Destroy | 7-day TTL, or `byteport preview remove` |

## Previews and pull requests

`byteport preview push` automatically detects an open PR and:
- Comments the preview URL on the PR
- Updates the comment on subsequent pushes
- Destroys the preview when the PR is closed/merged

## Mobile testing

Previews are also accessible from the BytePort mobile companion:
```bash
# Scan the QR code printed after `preview push` to open the
# preview in the mobile app — no app-store deploy required.
```

## Cleanup

```bash
# Remove this preview
byteport preview remove

# Remove all expired previews (manual cleanup)
byteport preview purge --expired
```

## Related

- Cookbook category: **DX**
