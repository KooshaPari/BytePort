# `backend/` — two Go modules live here

**Read this before building, testing, releasing, or debugging the backend.**

`backend/` contains **two independent Go modules**. They are not one server with
two packages; they have separate `go.mod` files, separate `models/` packages
that have drifted apart, and separate route surfaces.

| | **`backend/byteport/`** | **`backend/`** |
|---|---|---|
| Status | **Canonical** — the shipped backend | **Not reachable** by any client |
| Go module | `byteport` | `github.com/byteport/api` |
| First commit | 2024-11-23 (49 commits, active) | 2026-05-31 (2–4 commits per path) |
| Routes | `/authenticate`, `/health`, `/projects`, `/instances`, `/deploy`, `/login`, `/signup`, `/link`, `/user/:id/creds`, `/metrics`, `/github/*` | `/api/v1/*` |
| Default port | `8081` (`resolvePort()`: `PORT` → `BYTEPORT_API_PORT` → 8081) | `8090` |
| Auth | Own session/password flow | WorkOS AuthKit only |
| Built by | `Dockerfile` (`cd backend/byteport && go build -o /byteport-server`) | nothing |
| Models | `backend/byteport/models/` | `backend/models/` |
| SQLite driver | `github.com/glebarez/sqlite` (pure Go) | `gorm.io/driver/sqlite` |

## Which one to use

**`backend/byteport/` is the backend.** The desktop app talks to it and nothing
in the app or its frontend calls `/api/v1`:

```sh
cd backend/byteport
DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer \
SDKROOT=/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk \
CGO_ENABLED=1 go build -o /tmp/byteport-server .   # CGO optional; driver is pure Go
DATABASE_URL= PORT=8081 /tmp/byteport-server
```

The port is not arbitrary: the app fetches a fixed `http://localhost:8081`
(`API_PORT` in `frontend/web/src/lib/api.ts`) and the Tauri shell probes
`/health` there.

## Why this file exists

The split already caused a real mistake: a coordinator built and tested the
**wrong** module and got a misleading result — `403`/`404` from the outer module
where the canonical one would have answered. Because the outer module also
bound `8081`, a listener appeared to be up while serving an API the app never
calls, and `/api/v1/health` returned `200` — a healthy-looking process talking to
nobody.

Two guards are now in place: the outer module defaults to `8090`, and this file
records which is which.

## What is unresolved

Whether `backend/` should be **finished and adopted** (the WorkOS AuthKit
migration, with the frontend cut over) or **retired** is a product decision that
is not recorded anywhere in the repository. This file does not decide it. It
only removes the operational ambiguity: while both exist, the canonical one is
`backend/byteport/`, and the other is not wired to anything.

The full analysis, the drift evidence between the two `models/` packages, and
the options are in `docs/operations/two-backends.md`.
