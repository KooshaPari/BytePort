# Audit A3 — Stale Branches in BytePort

- **Unit:** A3 — List stale branches (>30d no commit)
- **Epic:** A — Hygiene garden & branch slim
- **Date:** 2026-06-24
- **Auditor:** koosh (compute-infra)
- **Type:** audit
- **Repo:** BytePort (KooshaPari/BytePort)

## Branch Inventory

**Total remote branches (excluding origin, origin/HEAD, origin/main):** 21

### STALE Branches (>30 days without commit)

| Branch | Last Commit | Days Stale | Author | Suggested Action |
|--------|------------|-----------|--------|-----------------|
| `ci/add-golangci-lint` | 2026-05-06 | 49d | Phenotype Agent | Delete (closed CI branch) |
| `ci/go-workflow` | 2026-05-06 | 49d | Phenotype Agent | Delete (closed CI branch) |
| `cve-sweep-high` | 2026-04-25 | 60d | Forge | Review & merge findings or delete |
| `docs/byteport-sladge-current` | 2026-05-06 | 49d | Phenotype Agent | Close/finalize doc branch |
| `docs/byteport-sladge-pem-current` | 2026-05-06 | 49d | Phenotype Agent | Close/finalize doc branch |
| `docs/sladge-badge` | 2026-04-29 | 56d | Codex | Close/finalize doc branch |
| `fix/add-pem-gitignore` | 2026-05-06 | 49d | Phenotype Agent | Delete (fix already landed or irrelevant) |
| `fix/codeql-apilink-ssrf` | 2026-04-26 | 59d | Forge | Review & merge or delete |
| `fix/deps-cve-2026-04-27` | 2026-04-27 | 58d | Codex | Review & merge or delete |
| `fix/nvms-parser-cleanup` | 2026-04-27 | 58d | Codex | Review & merge (Epic B dependency) |

**Stale count: 10 out of 21 (48%)**

### ACTIVE Branches (<=30 days)

| Branch | Last Commit | Days | Author |
|--------|------------|------|--------|
| `chore/governance-skeleton-2026-06-16` | 2026-06-16 | 8d | Phenotype Agent |
| `chore/manifest-fix` | 2026-06-05 | 19d | Phenotype Agent |
| `chore/slsa-build-2026-06-16` | 2026-06-16 | 8d | Cursor Agent |
| `ci/security-scan-2026-06-20` | 2026-06-20 | 4d | Phenotype Agent |
| `dependabot/cargo/major-2ceda2280c` | 2026-06-22 | 2d | dependabot[bot] |
| `dependabot/cargo/minor-and-patch-8468969804` | 2026-06-22 | 2d | dependabot[bot] |
| `dependabot/github_actions/major-6df859f817` | 2026-06-15 | 9d | dependabot[bot] |
| `dependabot/github_actions/minor-and-patch-cfdd69f9fc` | 2026-06-15 | 9d | dependabot[bot] |
| `dependabot/npm_and_yarn/frontend/web/npm-5453435d43` | 2026-06-10 | 14d | dependabot[bot] |
| `docs/a21-readme-workstate` | 2026-06-24 | 0d | KooshaPari |
| `fix/byteport-changelog-hygiene` | 2026-06-08 | 16d | Phenotype Agent |
| `t12-devcontainer-ci` | 2026-06-21 | 3d | KooshaPari |

## Dogfood Fixes Found

### 1. `Cargo.toml` — `resolver = "3"` is invalid

The workspace `Cargo.toml` used `resolver = "3"` which is not a valid resolver setting in stable Rust (valid values: `"1"` or `"2"`). This caused ALL grade.sh checks to fail (score 0/17).

**Fix applied:** Changed `resolver = "3"` to `resolver = "2"`.

## New DAG Units Suggested

1. **A11-BR** — Close/delete stale CI branches (`ci/add-golangci-lint`, `ci/go-workflow`)
2. **A12-BR** — Close/delete stale doc branches (`docs/byteport-sladge-current`, `docs/byteport-sladge-pem-current`, `docs/sladge-badge`)
3. **A13-BR** — Close/delete or merge stale fix branches (`fix/codeql-apilink-ssrf`, `fix/deps-cve-2026-04-27`, `fix/nvms-parser-cleanup`)
4. **A14-BR** — Clean up `fix/add-pem-gitignore` (stale, 49d)
5. **A15-BR** — Review `cve-sweep-high` findings and close

## Grade Summary

| Metric | Value |
|--------|-------|
| Score | 0 / 17 (0%) |
| Grade | F |
| Root cause | `resolver = "3"` fixed; re-run needed |
| Coverage baseline | N/A (all checks failed) |

## Coverage Delta vs main

Grade.sh could not complete due to the Cargo.toml resolver issue. Coverage delta computation deferred to re-run after `resolver = "2"` fix.

## Files Changed

```
M Cargo.toml                  (1 line — resolver "3" → "2")
A docs/audits/A3-stale-branches-2026-06-24.md  (audit report)
```

## Cross-repo Impact

- `phenotype-registry/projects/BytePort.json` — No architecture change, no update needed
- `DOMAIN_ROLES.md` — BytePort remains a CLI tool (no role change)
- `LANGUAGE_PLACEMENT.md` — Unchanged (Rust primary)
