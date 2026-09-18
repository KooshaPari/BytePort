# Jcode session handoff — resume on kooshapari-desk

**Session:** jcode, 2026-09-18 (Pacific). Repo under work: **BytePort**.
**Audience:** the Jcode session resuming this work on `kooshapari-desk`.
**Retrieval:** this file is committed and pushed, so `git clone` on the desktop
is enough. Everything else that only exists on the laptop is listed in §5.

---

## 1. What this session was doing

Two things, in this order:

1. **Fix BytePort's white window** (operator: "white screen", then "something in
   tauri broken, eval how we fixed in phenofabric").
2. **Fix the "naive 2023" UI** once it rendered (operator: "still ugly / my naive
   ui from 2023"), then **review the phenotype-canonicalization packet**.

All three are done. The operator judged the UI "good enough to move to other
steps" before I stopped.

## 2. Repo state

| Field | Value |
|---|---|
| Repo | `KooshaPari/BytePort` (public) |
| Remote | `git@github.com:KooshaPari/BytePort.git` |
| Branch | `main` |
| HEAD | `dfa27d63` |
| Unpushed | **0** |
| Laptop path | `/Users/kooshapari/CodeProjects/Phenotype/repos/wt-byteport-tauri-20260911` |

Despite the `wt-` prefix this is the **main** working tree, not a linked worktree.

**12 commits this session**, oldest first:

| Commit | What |
|---|---|
| `8415889e` | Declare `[features] custom-protocol = ["tauri/custom-protocol"]` |
| `00ff4d8b` | Makefile + README: stop instructing builds that produce white-window binaries |
| `6f512fd9` | CSP `connect-src` fix + `getBaseUrl()` `platform()` guard |
| `c10cc547` | Handoff + portable boot harness (`scripts/verify-frontend-boot.py`) |
| `373e58ad` | GORM driver selected from `DATABASE_URL` (was hardcoded postgres) |
| `de1ec6b2` | Canonical port 8081 + allow the macOS Tauri origin (`tauri://localhost`) |
| `daf03a74` | Handoff: two-backend split, CORS, Go SDK trap |
| `037c0308` | Token fix, graphite/teal palette, `src/lib/components/ui/*`, `src/lib/api.ts` |
| `b1b233ff` | Handoff: UI redesign + the token bug behind it |
| `9c349e4a` | Auth/entry views rebuilt |
| `511d9bd7` | `app.css` import order: stop shipping a stylesheet with no utilities |
| `9b826bbe` | Application shell |
| `ec1d437c` | Monitor + settings views |
| `fc29f968` | Projects + instances views and dialogs |
| `78e63b1e` | Entry route hands off to the shell; `Input` supports `bind:value` |
| `db7c7274` | svelte-check fix in the shell icon binding |

(16 rows; several were worker commits. `git log --oneline` is authoritative.)

Also written but **not** in the repo: the canonicalization review, §6.

## 3. The white screen: three layered causes

Reported symptom was one bug; it was three.

1. **Missing `custom-protocol` feature.** `src-tauri/Cargo.toml` lacked it, so
   every release binary loaded `devUrl` (`http://localhost:5173`) instead of
   embedded assets. With no Vite server running, WebKit failed with `-1004` and
   the window stayed white.
2. **CSP blocked the app's own backend.** `connect-src` omitted
   `http://localhost:8081`, so the packaged frontend could not reach its backend
   on any platform.
3. **`platform()` threw.** `getBaseUrl()` called it from
   `@tauri-apps/plugin-os`, which reads
   `window.__TAURI_OS_PLUGIN_INTERNALS__`. The Rust shell registers only
   `tauri-plugin-log`, so that global is `undefined` → `TypeError` → the promise
   rejected → `initializeUser` never ran.

PhenoFabric's earlier white-screen fix (`53079f6f`) was **unrelated** — that was
SSO iframes blocked by `X-Frame-Options`.

## 4. The "ugly" UI: what was actually wrong

Not primarily taste. `app.css` had `@config` on line 1 followed by
`@import 'tailwindcss'`. CSS requires `@import` to precede other rules, so it was
invalid and **silently dropped**:

| app.css order | Built CSS | Utilities | Preflight |
|---|---|---|---|
| `@config` then `@import` | 8,096 B | **none** | **none** |
| `@import` first | 44,272 B | `.flex`, `.grid`, `.p-4`… | present |

So `class="flex p-4 rounded-lg"` did nothing on every screen, and the build
exited 0 throughout. Separately, `--border`/`--input`/`--ring` held hex while
`tailwind.config.ts` consumed them as `hsl(var(--x))`, so
`border-color: hsl(#3f4948 / 1)` was invalid and **every border fell back to
`currentColor`**.

Then: palette replaced (Material 3 → graphite + one teal accent, keeping the
`dark*` names so ~200 existing classes upgraded with no markup churn), shared
primitives added under `src/lib/components/ui/`, `src/lib/api.ts` as the single
backend-URL source, all views rebuilt, and the entry route rewritten (it rendered
a **second** sidebar and a `<img src="/src/assets/img/byte.png">` that never
resolves).

## 5. Verified vs UNKNOWN

**Verified by behaviour, not inspection:**

- Webview commits the document: `didCommitLoadForFrame` (was
  `didFailProvisionalLoad ... -1004`).
- Real backend round trip from the installed app:
  `204 OPTIONS /authenticate` then `401 GET /authenticate`.
- `npm run build` exit 0; built CSS 44,272 B with preflight and utilities.
- Workers reported `npx svelte-check` **0 errors / 0 warnings** across 2,721
  files, and Playwright DOM harnesses **34/34** (auth) and **39/39** (projects)
  with zero console errors.

**UNKNOWN — do not inherit as done:**

- **Visual quality.** No screenshots were ever taken (the capture restriction
  forbids them), so nobody in the loop saw the rendered window. The operator
  called it "good enough"; that is the only visual judgement that exists.
- **Endpoints other than `/authenticate`** were never exercised from the app.
- Backend saves/loads are contract-verified only, never integration-tested.

## 6. Canonicalization review (done, outside the repo)

Reviewed `/Users/kooshapari/Downloads/phenotype-canonicalization`; wrote
`REVIEW-2026-09-18.md` **beside the packet** (it is not in any repo, so it does
not travel with a clone).

- Packet verified: structure `VALID`, JSON-schema `VALID`, reference-tool tests
  **40/40**, file inventory complete, ships `SHA256SUMS`. Its claims
  "Tracera pins TS 5.9.3" and "`phenotype-tooling` 404" both re-confirmed.
- **Main finding:** the packet's declared authority, `GLOBAL_HANDBOOK.md`, exists
  in **21 copies with 3 divergent contents**. The `6c85e9da` variant (PhenoShared,
  pheno, phenotype-omlx) knows `docs-5` superseded `docs-3`; the `8ab46839`
  variant in **15 repos** still names `docs-3` as the live atlas. Every copy says
  the same sentence: *"Supersedes all prior loose policy documents. This is the
  single canonical reference…"*
- Recommendation left with the owner: reconcile centrally, then delete the
  redundant copies **or** replace them with a pointer plus content hash. I did not
  delete anything.
- The packet's WP02 (Tracera repeated-`-p` typecheck hazard) is **already fixed**
  by Tracera `427ffd3af`.

## 7. Environment traps (these cost real time; do not rediscover them)

- **Bare `cargo build --release` produces a white-window binary.**
  `frontend/web/src-tauri` is a workspace member, so it compiles the GUI without
  `custom-protocol`. Use `cargo tauri build` or `make build-app`.
- **Asset-only changes do not trigger a relink.** After changing frontend files,
  `touch frontend/web/src-tauri/src/main.rs` or the old assets stay embedded.
- **Go builds need Xcode's SDK pinned.** CommandLineTools SDK 27.0 ships `.tbd`
  files with an `arm64e.x1-macos` architecture its own linker rejects:
  ```bash
  DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer \
  SDKROOT=/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk \
  CGO_ENABLED=1 go build -o /tmp/out .
  ```
- **`@import 'tailwindcss'` as a bare specifier fails** with
  `ENOENT .../frontend/web/tailwindcss`; Vite joins bare specifiers to the
  importing file's directory, so `app.css` uses a resolvable relative path.
- **Two backends exist.** `backend/` (module `github.com/byteport/api`,
  `/api/v1`, default port 8080) and `backend/byteport/` (module `byteport`,
  `/authenticate`, port 8081). **Only the second serves the app.** There are two
  `models/` packages that have drifted.
- **Port 8080 is already taken** on the laptop by `sl-daemon`, which is further
  evidence 8081 is the intended value.
- **Concurrent `vite build` races `.svelte-kit/output`** and can fail with ENOENT.
  Serialise builds; clear `.svelte-kit/output` if it happens.
- Cargo was run with `CARGO_NET_OFFLINE=true`; `vendor/clap-ext` is vendored
  because its upstream 404s. `tauri-plugin-os` is neither cached nor vendored.

## 8. Only on the laptop (needs SSH from the desktop)

| Item | Detail |
|---|---|
| Laptop | `Kooshas-Laptop.local`, macOS 27.0, arm64 |
| Tailscale (use this) | **`100.112.14.98`** (`kooshas-laptop`) |
| LAN | `192.168.1.23` |
| Remote login | `com.openssh.sshd => enabled`; no listener seen on :22 (launchd socket-activates it, so it starts on connect) |
| Desktop peer | `kooshapari-desk` — **Windows**, Tailscale `100.96.135.160`, LAN `192.168.1.159` |
| Repo checkout | `/Users/kooshapari/CodeProjects/Phenotype/repos/wt-byteport-tauri-20260911` |
| Installed app | `/Applications/Byteport.app` (rebuilt + ad-hoc signed this session) |
| Running backend | `/Users/kooshapari/.jcode/scratch/bp-cors2` on :8081 |
| Boot harness | `scripts/verify-frontend-boot.py` — **in the repo**, so it travels |
| no-mistakes daemons | PIDs 38900 (`~/.airlock/bin`, root `~/.no-mistakes`) and 927 (`~/.local/bin`) |
| Airlock socket | `~/.airlock/socket` → symlink → `~/.no-mistakes/socket` |

The desktop is Windows, so laptop paths (`/Users/kooshapari/...`) do not apply
there; clone from GitHub and expect a different checkout root.

## 9. Open work

| Priority | Item |
|---|---|
| High | Resolve the SQLite-vs-Postgres contradiction. Docs promise `file:./byteport.db`; models use `uuid DEFAULT gen_random_uuid()` and `jsonb`, so migration fails with `near "(": syntax error`. Check **both** `backend/models/` and `backend/byteport/models/`. |
| High | Reconcile the two backends, or document the split deliberately |
| High | The handbook drift in §6 — 15 repos still name `docs-3` as live. **Needs the owner's delete/redirect decision.** |
| Medium | Migrate to `@tailwindcss/vite` to drop the explicit `node_modules` import path |
| Medium | Add a backend test for the `api_key` casing bug: `models.AIProvider` tags the field `json:"api_key"` and the handler looks up the map key `"openai"`, but the old UI sent `apiKey`/`openAI`, so the OpenAI key always validated as empty. Two workers found this independently. |
| Medium | Register `tauri-plugin-os` in `src-tauri/src/lib.rs` once network allows; the `try/catch` then becomes belt-and-braces |
| Medium | `frontend/web/src-tauri/benches/ipc.rs` is broken (`criterion` and `app_lib::ipc::IpcEnvelope` are not wired); `cargo check --all-targets` fails |
| Medium | ~50 pre-existing dirty files in the BytePort tree (a PII-redaction sweep: `kooshapari` → `<REDACTED>`). **Not from this session — do not discard, do not commit.** |
| Medium | 27 repos in `~/CodeProjects/Phenotype/repos` have unpushed commits, **259 total** (largest: `portage` 161, `koosha-phenotype` 33, `khostty` 19). Laptop-only; verify before pushing. |
| Security | A Cloudflare API token was exposed in `PhenoInfra` history and stripped with BFG. **It was still valid when checked — rotate it.** |

## 10. Working agreements this session followed

- The visual gate is a **restriction, not a preference**: no desktop,
  whole-display or arbitrary-window screenshots; only an isolated agent-owned
  process with verified ownership may ever be captured. Otherwise leave the gate
  UNKNOWN rather than substituting a proxy. That is why §5 says UNKNOWN.
- Distinguish observed inventory / candidate / local validation / hosted
  validation / delivered acceptance. A clean `git status` is not readiness.
- Historical evidence keeps its observation date; it must not silently become a
  fresh pass.
- No force-push, no history rewrite, no `git reset --hard`. Commits carry
  `tx-agent` / `tx-validated` trailers.
- One agent, one repo. Prefer primitives and shared modules over per-view copies.
- When spawning parallel workers on one repo, assign **exclusive file
  ownership**. The harness `Files:` attribution includes **reads**, not just
  writes — check `git` before accusing a worker of touching shared files. I got
  this wrong once and had to retract it.

## 11. First three actions

1. `git clone git@github.com:KooshaPari/BytePort.git`, confirm HEAD is
   `dfa27d63` and that `git log --oneline @{u}..HEAD` is empty.
2. `cargo tauri build`, then run the §"boot harness" from
   `scripts/verify-frontend-boot.py` against the fresh binary. Expect a
   `/authenticate` hit; if absent, check the CSP `connect-src` and `platform()`
   first, since both failed silently before.
3. Pick up the High-priority items in §9. The highest-leverage is the handbook
   drift, but it needs the owner's decision before anything is deleted.
