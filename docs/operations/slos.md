# Service Level Objectives (SLOs)

> **Last reviewed:** 2026-06-30
> **Owner:** @KooshaPari
> **Review cadence:** Quarterly (next review: 2026-09-30)

## Scope

BytePort provides two main service surfaces:

1. **Go API server** (`backend/`) — HTTP REST endpoints consumed by the Astro frontend and external CLI callers.
2. **NVMS deploy pipeline** (`backend/nvms/projectManager/`) — long-running deploy orchestration triggered by `POST /deploy`.

For the Rust crates (`byteport-cli`, `byteport-transport`, etc.) the "service" is the library/CLI API surface; SLOs are expressed as build + test success rate rather than uptime.

---

## SLO Table

| Service | SLI | Target | Measurement window | Error budget | Burn-rate alert |
|---------|-----|--------|--------------------|--------------|-----------------|
| API server — availability | % of requests returning non-5xx | 99.5% | 30 days rolling | 3.6 h/month | Alert at 2x burn over 1 h |
| API server — p99 latency | p99 response time for `GET` routes | <= 500 ms | 7 days rolling | — | Alert if p99 > 1 s for 5 min |
| Deploy pipeline — success rate | % of `POST /deploy` calls that complete without error | 95% | 30 days rolling | 5% of deploys | Alert if 3 consecutive failures |
| Deploy pipeline — end-to-end latency | p95 wall-clock time from request to "deployed" status | <= 120 s | 7 days rolling | — | Alert if p95 > 240 s |
| Rust build / test CI | % of `cargo test --workspace` runs that pass on `main` | 99% | 30 days rolling | ~7 min/month | Alert on 2 consecutive failures |
| Go build / test CI | % of `go test ./...` runs that pass on `main` | 99% | 30 days rolling | ~7 min/month | Alert on 2 consecutive failures |

---

## Measurement Methodology

**API availability and latency**: Measured via Gin middleware request logging to stdout.
Structured log lines include `status`, `latency`, and `path`. When an OTel exporter
is wired (tracked in L5 observability work), these will flow to a metrics backend.
Until then, availability is approximated from CI green rate and manual spot checks.

**Deploy pipeline**: Success/failure is recorded in the `deployments` SQLite table
(`backend/database.db`) via `project.AppendDeploy` and status update calls in
`backend/nvms/projectManager/deploy.go`. Query example:

```sql
SELECT
  SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) * 1.0 / COUNT(*) AS error_rate
FROM deployments
WHERE created_at >= datetime('now', '-30 days');
```

**CI pass rate**: GitHub Actions run history for `ci.yml` and `tier2-coverage-gate.yml`
on the `main` branch. Tracked via the Actions tab; no automated burn-rate tooling yet.

---

## Current Gaps (tracked)

- No automated burn-rate alerting — manual review only until OTel metrics are wired (L5 work).
- No latency histogram export — p99/p95 values are estimated from logs, not measured precisely.
- SLO compliance not yet surfaced in the BytePort dashboard (future L14 work).
- Runbooks for common failure modes are pending (L27 work).

---

## Related Docs

- `docs/operations/journey-traceability.md` — user journey to SLI mapping
- `ARCHITECTURE.md` — service topology
- `SECURITY.md` — availability-relevant security controls
- `README.md` — quickstart and project overview
