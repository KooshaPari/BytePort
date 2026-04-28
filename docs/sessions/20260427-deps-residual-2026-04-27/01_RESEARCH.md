# Research

Confirmed open Dependabot alerts via GitHub API:
- `Cargo.lock` -> `rand` and `glib`
- `frontend/web/src-tauri/Cargo.lock` -> `rand` and `glib`
- `frontend/web/yarn.lock` -> `cookie` and `prismjs`
- `frontend/web/package-lock.json` -> `cookie` and `prismjs`
- `.github/frontend/package-lock.json` -> `cookie`

Cargo graph findings:
- `rand 0.7.3` comes from `tauri-utils 2.8.3 -> selectors 0.24.0 -> phf_codegen 0.8.0`.
- `glib 0.18.5` comes from the published `tauri 2.10.3` / `gtk 0.18.2` / `webkit2gtk 2.0.2`
  chain.
- The published crate versions available from crates.io do not expose a clean manifest-only
  upgrade path to `glib 0.20.0` from this workspace.
