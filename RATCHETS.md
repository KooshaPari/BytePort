# BytePort Quality Ratchets

> Pillars addressed: L11 (quality gates), L38 (ratchets), L32 (test infrastructure).
> Last updated: 2026-06-30

Quality ratchets prevent metric regression. Once a threshold is met and the ratchet
is bumped, the CI gate enforces the new floor — coverage cannot drop below it without
a deliberate ratchet update committed to `main`.

## Coverage Ratchets

| Tier | Enforced by | Current threshold | Next target | Workflow |
|------|-------------|-------------------|-------------|----------|
| Rust workspace | `cargo llvm-cov --fail-under-lines` | 50% lines | 60% lines | `fr-coverage.yml` |
| Rust lib (`byteport-transport`) | `cargo llvm-cov --fail-under-lines` | 71% lines | 75% lines | `tier2-coverage-gate.yml` |
| Go backend (`backend/`) | `go tool cover` threshold check | 50% | 70% | `fr-coverage.yml` |
| Go framework (all packages) | `go tool cover` threshold check | 70% | 75% | `tier2-coverage-gate.yml` |
| Frontend / service (Vitest) | vitest coverage json parse | 60% lines | 70% lines | `tier2-coverage-gate.yml` |

### How to bump a ratchet

1. Confirm the new baseline passes locally: run the relevant coverage command and verify the output.
2. Update the threshold value in the workflow file (column "Enforced by").
3. Update the "Current threshold" row in this table.
4. Commit both changes together with message: `chore(ratchet): bump <tier> coverage to <N>%`.

## Lint Ratchets

| Check | Tool | Enforcement | Threshold |
|-------|------|-------------|-----------|
| Rust clippy warnings | `clippy` | `clippy.toml` `deny` list | 0 warnings (deny = true) |
| Go vet + staticcheck | `golangci-lint` | `golangci.yml` | 0 errors |
| TypeScript | ESLint | `jest.config.js` | 0 errors |

## Complexity Ratchet

Not yet enforced. Target: add `cargo-cranky` or equivalent for Rust cyclomatic complexity
and `gocyclo` for Go with a per-function limit of 15. Track in L11 pillar work.

## Mutation Testing

Not yet enabled. Target: `cargo-mutants` for Rust lib crates on nightly or weekly schedule.
Track in L11 pillar work.

## How ratchets relate to CI gates

- `fr-coverage.yml` — runs on every PR; enforces the "Current threshold" floors above.
- `tier2-coverage-gate.yml` — runs on every PR + push to main; enforces Rust lib and Go framework floors.
- Both gates block merge if thresholds are not met.
- The `.coverage-baseline` file at repo root records the last committed baseline for
  tooling that reads it; update it in sync with ratchet bumps.
