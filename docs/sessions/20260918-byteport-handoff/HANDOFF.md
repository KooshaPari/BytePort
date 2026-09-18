# BytePort — Session Handoff

**Author:** jcode (session `seedling`), 2026-09-18
**Audience:** the next agent session that owns this repo and all future BytePort work.
**Status:** white-window defect resolved and verified. Three commits pushed. See
"Verified vs unverified" before trusting anything here.

---

## 1. Repo identity

| Field | Value |
|---|---|
| GitHub | `KooshaPari/BytePort` (public) |
| Remote | `git@github.com:KooshaPari/BytePort.git` |
| Default branch | `main` |
| HEAD at handoff | `6f512fd9` (2026-09-18 01:38:15 -0700) |
| Local path (this machine) | `/Users/kooshapari/CodeProjects/Phenotype/repos/wt-byteport-tauri-20260911` |

Despite the `wt-` prefix this is the **main** working tree, not a linked worktree
(`git worktree list` shows a single entry). Do not be misled by the name.

Clone fresh on the desktop:

```bash
git clone git@github.com:KooshaPari/BytePort.git
```

## 2. What BytePort is

A Tauri 2 desktop deployment app: Rust shell (`frontend/web/src-tauri`),
SvelteKit + Vite frontend (`frontend/web`), Go backend
(`backend/byteport`, Gin, listens on `localhost:8081`). It also ships a
headless CLI (`crates/byteport-cli`) with `version`, `transport status`,
`transport check`, and `deploy --target <docker|local|nanovms|spin>`.

---

## 3. Commits pushed this session

| Commit | Purpose |
|---|---|
| `8415889e` | Declare `[features] custom-protocol = ["tauri/custom-protocol"]` in `frontend/web/src-tauri/Cargo.toml` |
| `00ff4d8b` | Makefile + README: stop instructing builds that produce white-window binaries |
| `6f512fd9` | CSP `connect-src` fix + `getBaseUrl()` `platform()` guard |
| `c10cc547` | This handoff + `scripts/verify-frontend-boot.py` |
| `373e58ad` | GORM driver selected from `DATABASE_URL` instead of hardcoded postgres |
| `de1ec6b2` | Canonical port 8081 + allow the macOS Tauri origin (`tauri://localhost`) |

All commits are on `origin/main`. `git log --oneline @{u}..HEAD` is empty.

**Trap for reviewers:** this working tree also carries an unrelated,
pre-existing PII-redaction sweep (`kooshapari` → `<REDACTED>`, 50+ files across
`docs/`, root `*.md`, `backend/server.go`, `backend/bytebridge/.../spin.toml`).
It is **not** part of this session's work. Stage by explicit path; never
`git add -A`.

---

## 2b. Backend architecture — there are TWO backends

This is the single most confusing thing in the repo. Both live under `backend/`:

| | `backend/` | `backend/byteport/` |
|---|---|---|
| Go module | `github.com/byteport/api` | `byteport` |
| Entry point | `main.go` → `server.go` | `main.go` |
| Routes | `/api/v1/*` | `/authenticate`, `/api/...` |
| Port | `defaultPort` 8081 (was 8080) | `resolvePort()` → 8081 |
| Models | `backend/models/` | `backend/byteport/models/` |
| Used by the desktop app? | **No** | **Yes** |

The SvelteKit frontend fetches `http://localhost:8081/authenticate`, which only
`backend/byteport` serves. Verify any backend change against the right module —
earlier in this session a test was run against the wrong one and produced a
misleading 403/404.

`backend/models/` and `backend/byteport/models/` are **separate packages that
have drifted**. Do not assume a fix in one applies to the other.

### Environment trap: Go builds need a pinned SDK

`go build` fails in `backend/models` (any package pulling CGO, e.g. the SQLite
driver) because the CommandLineTools SDK 27.0 `.tbd` files carry an
`arm64e.x1-macos` architecture the CommandLineTools linker rejects:

```
libresolv.9.tbd:4:20: error: unknown architecture arm64e.x1-macos
```

Pin Xcode's SDK instead:

```bash
DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer \
SDKROOT=/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk \
CGO_ENABLED=1 go build -o /tmp/byteport-api .
```

This links with only "built for newer macOS version" warnings.

### Root cause of the reported "white screen"

`frontend/web/src-tauri/Cargo.toml` was missing the Tauri 2 template's
`custom-protocol` feature declaration. Without it, a release binary loads
`devUrl` (`http://localhost:5173`) instead of the embedded `frontendDist`
assets. With no Vite dev server running, WebKit failed with `-1004` and the
window stayed white.

Two further defects were found while closing the verification loop:

1. **CSP blocked the app's own backend.** `app.security.csp` `connect-src`
   omitted `http://localhost:8081`, so the packaged frontend could not reach
   its local backend on any platform.
2. **`platform()` threw.** `getBaseUrl()` calls `platform()` from
   `@tauri-apps/plugin-os`, which reads
   `window.__TAURI_OS_PLUGIN_INTERNALS__.platform`. The Rust shell registers
   only `tauri-plugin-log`, so that global is `undefined` → `TypeError` →
   the promise rejected → `initializeUser` never ran.

---

## 4. Verified vs unverified

**Verified with behavioural evidence (not inspection):**

- Bundle assets load inside the real Tauri webview. Probe server on 8081
  observed the webview complete a genuine request:
  ```
  OPTIONS /authenticate origin=tauri://localhost
  GET /authenticate ua=Mozilla/5.0 (Macintosh; Intel Mac OS X ...)
  ```
  This proves SvelteKit booted, the root route mounted, `onMount` ran, and
  `getBaseUrl()` resolved. Before the fixes the webview logged
  `didFailProvisionalLoad ... code=-1004`; after, `didCommitLoadForFrame`.
- No failed subresource loads after the load commits.
- `make build` and `cargo tauri build` both produce working binaries.
  `cargo tauri build` emits `target/release/bundle/macos/Byteport.app` and
  `target/release/bundle/dmg/Byteport_0.1.0_aarch64.dmg` (~3.8 MB).
- `cargo build --release --workspace --exclude app` succeeds (the README's
  documented non-GUI build).

- **Backend integration against the real backend (`backend/byteport`).** With
  that module running on 8081 and the installed app launched, the backend's own
  request log recorded:
  ```
  204 | OPTIONS "/authenticate"   <- preflight accepted (403 before the CORS fix)
  401 | GET     "/authenticate"   <- real request followed; 401 = not logged in
  ```
  A genuine request/response cycle between the packaged desktop app and the real
  Go backend, not a stub.
- **CORS fix, including a failure caught by testing.** The first attempt -
  adding `tauri://localhost` to `AllowOrigins` - *panicked at startup* because
  `gin-contrib/cors` rejects non-http(s) schemes. The shipped fix uses
  `AllowOriginFunc`. Recorded because the naive approach looks correct and
  breaks the server.

**NOT verified — treat as UNKNOWN:**

- **Visual appearance.** The capture restriction forbids screenshots, so
  nobody has confirmed the window *looks* correct. Behaviour is proven;
  pixels are not. Ask the operator to eyeball it.
- **Endpoints other than `/authenticate`.** Projects, git search, deployments
  and the WorkOS auth flow were not exercised end to end.
- **The SQLite path.** See open work: the documented SQLite default now selects
  the right driver, but migration then fails because the models use
  PostgreSQL-only types.
- **Signing for distribution.** `bundle.macOS` is `{}`, so Tauri produces
  only the Rust linker's ad-hoc signature and no bundle `_CodeSignature`.
  I ad-hoc signed the installed copy manually (`codesign --force --deep
  --sign -`). Notarisation/Developer ID is unaddressed.
- **Windows/Linux builds.** `bundle.targets` lists deb/rpm/appimage/msi/nsis,
  but only the macOS aarch64 path was exercised.

---

## 5. How to rebuild (exact commands)

```bash
cd frontend/web && npm install        # node 26.8.1 / npm 11.19.0 both work

# Canonical release path (builds frontend, enables custom-protocol, bundles)
cargo tauri build
# or: make build-app

# Rust workspace incl. a correctly-configured GUI binary
make build            # runs the frontend build, then
                      # cargo build --release --workspace --features app/custom-protocol

# Backend/CLI only, GUI excluded
cargo build --release --workspace --exclude app
```

### Traps that will silently reintroduce the white window

`frontend/web/src-tauri` **is** a workspace member. Therefore:

- `cargo build --release` (bare, at the root) compiles the GUI **without**
  `custom-protocol` → white-window binary.
- `cargo build --release -p app` → same problem.
- A bare cargo rebuild **does not re-embed changed frontend assets**.
  `tauri-build` did not relink when only `frontend/web/build/**` changed; I
  had to `touch frontend/web/src-tauri/src/main.rs` to force re-embedding.
  If you change frontend code and rebuild, verify the assets actually landed.

### Offline constraint

Builds were done with `CARGO_NET_OFFLINE=true`. `vendor/clap-ext` is
vendored because its upstream GitHub repo 404s. `tauri-plugin-os` is **not**
in the cargo cache and not vendored, which is why the `platform()` problem
was fixed defensively in the frontend rather than by registering the plugin.
If you have network, registering `tauri-plugin-os` is the fuller fix.

---

## 6. Verification harness (reusable)

A probe server that proves the frontend booted:

- Script: `scripts/verify-frontend-boot.py` (in this repo, portable)
- Behaviour: listens on 8081, logs every request to
  `/tmp/byteport-boot-probe.log`, answers `/authenticate` with JSON.
- It must reflect `Access-Control-Request-Headers` (not `*`) — wildcard
  `Allow-Headers` is invalid with `credentials: 'include'`, and the browser
  will send the preflight but never the GET.

Run it:

```bash
python3 scripts/verify-frontend-boot.py 8081 &
pkill -f 'release/byteport'
nohup ./target/release/byteport >/dev/null 2>&1 &
sleep 15 && cat /tmp/byteport-boot-probe.log
```

A hit on `/authenticate` means the UI booted. **No hit means investigate, not
that it is broken** — check CSP and `platform()` first.

WebKit-side evidence (what actually decided the root cause):

```bash
log show --last 5m --style compact --predicate 'process == "byteport"' \
  | grep -E 'didCommitLoadForFrame|didFailProvisionalLoad'
```

---

## 7. Machine-local state (needs SSH)

The desktop session will not inherit any of this:

| Item | Location / detail |
|---|---|
| Host | `Kooshas-Laptop.local`, macOS 27.0, arm64 |
| Tailscale IP | `100.112.14.98` (preferred for SSH from the desktop) |
| LAN IP | `192.168.1.23` |
| sshd | **UNKNOWN** — `lsof -iTCP:22` showed no listener and launchd could not be queried without sudo. Confirm before relying on SSH. |
| Probe server | `~/.jcode/scratch/bp-probe-server.py`, `bp-probe-hits.log` |
| Build binaries | `target/release/byteport`, `target/release/bundle/**` |
| Installed app | `/Applications/Byteport.app` — refreshed + ad-hoc signed this session |
| SSH keys | `~/.ssh/id-git`, `id_helioslite`, `oci_ampere_phenotype`, `push_key` |

### Git gate daemons (this machine only)

Two `no-mistakes` daemons run here:

- PID 38900 — `~/.airlock/bin/no-mistakes daemon run --root ~/.no-mistakes`
- PID 927 — `~/.local/bin/no-mistakes ... --root .../repos/.coordination/no-mistakes-pr7-nmhome`

Socket wiring: `~/.airlock/socket` → symlink → `~/.no-mistakes/socket`
(the real socket). If a push hangs, suspect this wiring.

`git push no-mistakes <branch>` routes through the gate; the daemon runs its
pipeline in a disposable worktree and only then pushes to origin. Six repos
have a `no-mistakes` remote configured.

---

## 8. Open work

| Priority | Item |
|---|---|
| High | Resolve the SQLite-vs-Postgres contradiction. `INSTALL.md`/`DEPLOYMENT.md`/`docker-compose.yml` promise `file:./byteport.db`, but the models use PostgreSQL-only types (`uuid DEFAULT gen_random_uuid()`, `jsonb`), so AutoMigrate fails with `near "(": syntax error`. Either make SQLite work (app-generated UUIDs, portable JSON) or correct the docs. **Check both** `backend/models/` and `backend/byteport/models/` — they have drifted. |
| High | Reconcile the two backends (`backend/` vs `backend/byteport/`). They are separate modules with different routes and duplicated, drifted models. Decide which is canonical and retire the other, or document the split deliberately. |
| High | Register `tauri-plugin-os` in `src-tauri/src/lib.rs` once network allows; then the `try/catch` in `+page.svelte` becomes belt-and-braces |
| High | Exercise endpoints other than `/authenticate` (projects, git search, deployments, WorkOS auth) from the packaged app |
| Medium | Set `bundle.macOS.signingIdentity` (or rely on `tauri-action` env) so bundles verify without manual ad-hoc signing |
| Medium | `frontend/web/src-tauri/benches/ipc.rs` is broken — uses `criterion` and `app_lib::ipc::IpcEnvelope`, neither of which is wired. `cargo check --all-targets` fails. |
| Medium | Confirm visual appearance with the operator (only they can see it) |
| Low | ~50 pre-existing uncommitted files in the working tree (PII-redaction sweep, workflow churn). **Not from this session — do not discard, do not commit.** |
| Low | 8 Dependabot alerts on `main` (2 high, 6 moderate), typical Tauri 2.x ecosystem |

### Security follow-up (not BytePort)

A Cloudflare API token was exposed in `PhenoInfra` history and was stripped
with BFG. **The token was still valid when checked — it should be rotated at
cloudflare.com.** Tracked here because the same operator owns both repos.

### Portfolio state at handoff (for context, not this repo's work)

| Repo | Branch | Ahead | Dirty |
|---|---|---|---|
| `sharecli` | main | 2 | 304 |
| `substrate` | main | 3 | 50 |
| `PhenoInfra` | main | 1 | 373 |
| `Tracera` | main | 1 | 27 |
| `HeliosLab` | main | 0 | 13 |

`substrate` is **not** published under `KooshaPari` on GitHub (404).

---

## 9. Working agreements the next session must honour

- The visual gate is a **restriction, not a preference**: no desktop,
  whole-display, or arbitrary-window screenshots; only an isolated
  agent-owned process with verified ownership may ever be captured. If
  isolation cannot be guaranteed, leave the gate UNKNOWN rather than
  substituting a proxy.
- Distinguish observed inventory / candidate / local validation / hosted
  validation / delivered acceptance. A clean `git status` is not readiness.
- Historical evidence keeps its observation date and must not silently become
  a fresh pass.
- No force-push, no history rewrite, no `git reset --hard`. Commits carry
  `tx-agent` / `tx-validated` trailers.
- One agent, one repo. BytePort scope only; RepoLedger is intentionally
  retired; ResearchLedger goes last; honour sponsor account-name holds.
- Do not commit the 46 pre-existing dirty files. Stage explicitly by path.

## 10. First three actions for the new session

1. `git clone git@github.com:KooshaPari/BytePort.git && git log --oneline -5`
   and confirm `de1ec6b2` is HEAD.
2. `cargo tauri build`, then run the §6 probe against the fresh binary. Expect
   a `/authenticate` hit. If absent, check CSP `connect-src` and `platform()`.
3. Run the **correct** backend module — `cd backend/byteport && go build` with
   the pinned SDK from §2b — and confirm the app reaches it (expect
   `204 OPTIONS /authenticate` then `401 GET /authenticate`). Do not test
   against `backend/` (the `/api/v1` module); it is not the app's backend.
