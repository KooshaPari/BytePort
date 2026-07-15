# BytePort SOC 2 Controls Mapping

This document maps BytePort's security controls to the SOC 2 Trust Services Criteria
(TSC).  Controls are categorised under **Security**, **Availability**, **Processing
Integrity**, **Confidentiality**, and **Privacy** (the five Trust Services Categories).

## 1. Security

| Criterion | Control ID | Implementation | Status |
|-----------|-----------|----------------|--------|
| CC1.1 Board oversight | GOV-01 | Product lead reviews security posture monthly | ✅ Implemented |
| CC2.1 Risk assessment | RISK-01 | `docs/sessions/` audit track records risk decisions | ✅ Implemented |
| CC3.1 Access control | AUTH-01 | JWT-based auth with expiring sessions; API key for MCP | ✅ Implemented |
| CC3.2 Logical access | AUTH-02 | Role-based permissions (admin / member / viewer) | ✅ Implemented |
| CC3.3 Physical access | PHYS-01 | Cloud-hosted (Vercel / Railway) — vendor SOC 2 relied upon | ✅ Vendor-managed |
| CC4.1 System monitoring | MON-01 | Prometheus metrics + Grafana dashboards (`observability/`) | ✅ Implemented |
| CC4.2 Anomaly detection | MON-02 | Circuit breakers (`internal/infrastructure/resilience/`) | ✅ Implemented |
| CC5.1 Change management | CI-01 | CI/CD via GitHub Actions with mandatory review | ✅ Implemented |
| CC5.2 Separation of environments | CI-02 | Dev / staging / prod via `docker-compose` profiles | ✅ Implemented |
| CC6.1 Incident response | IR-01 | DAST scanning (`dast.yml`) + SBOM (`sbom-supply-chain.yml`) | ✅ Implemented |
| CC6.2 Data backup | DR-01 | Postgres WAL + periodic snapshots via Supabase | ✅ Vendor-managed |
| CC7.1 Vulnerability management | VULN-01 | Weekly `pillar-audit.yml` + Dependabot + SBOM generation | ✅ Implemented |

## 2. Availability

| Criterion | Control ID | Implementation | Status |
|-----------|-----------|----------------|--------|
| A1.1 Capacity planning | CAP-01 | Bench workflow (`bench.yml`) measures throughput | ✅ Implemented |
| A1.2 Performance monitoring | MON-03 | Grafana dashboards track latency / error rate / throughput | ✅ Implemented |
| A1.3 Disaster recovery | DR-02 | Multi-region cloud deployment; health endpoints (`/healthz`, `/readyz`) | ✅ Implemented |
| A1.4 Business continuity | BCP-01 | Stateless API — horizontal scale via `docker-compose up --scale` | ✅ Implemented |

## 3. Processing Integrity

| Criterion | Control ID | Implementation | Status |
|-----------|-----------|----------------|--------|
| PI1.1 Completeness | INT-01 | Property-based tests (`deployment_rapid_test.go`) | ✅ Implemented |
| PI1.2 Accuracy | INT-02 | Contract tests (`internal/infrastructure/contract/`) | ✅ Implemented |
| PI1.3 Timeliness | INT-03 | SLI / SLO tracking via Prometheus metrics | ⚡ In progress (see SLA.md) |
| PI1.4 Authorised processing | INT-04 | Auth middleware validates every request | ✅ Implemented |

## 4. Confidentiality

| Criterion | Control ID | Implementation | Status |
|-----------|-----------|----------------|--------|
| C1.1 Encryption at rest | ENC-01 | Postgres transparent encryption (AES-256) via Supabase | ✅ Vendor-managed |
| C1.2 Encryption in transit | ENC-02 | TLS 1.3 for all HTTP endpoints; mTLS for mesh | ✅ Implemented |
| C1.3 Key management | KMS-01 | Secrets via environment variables; cosign for signing | ✅ Implemented |
| C1.4 Data classification | CLS-01 | All deployment artefacts tagged with sensitivity label | ✅ Implemented |

## 5. Privacy

| Criterion | Control ID | Implementation | Status |
|-----------|-----------|----------------|--------|
| P1.1 Notice | PRV-01 | Privacy notice in `README.md` | ⚡ Planned |
| P1.2 Consent | PRV-02 | Opt-in telemetry via `byteport-otel` crate | ✅ Implemented |
| P1.3 Data minimisation | PRV-03 | Only essential telemetry collected; user IDs are UUIDs | ✅ Implemented |
| P1.4 Retention & disposal | PRV-04 | Soft-delete with `deleted_at` timestamp on all entities | ✅ Implemented |

## Evidence & Artefacts

| Artefact | Location | Automation |
|----------|----------|------------|
| Pipeline logs | `.github/workflows/` | GitHub Actions |
| Vulnerability scan | `dast.yml` | Weekly OWASP ZAP |
| Dependency audit | `sbom-supply-chain.yml` | CycloneDX + cosign |
| Access logs | Cloud provider console | Vendor-managed |
| Incident response | `docs/sessions/` | Session logs |
| Security review | `pillar-audit.yml` | Weekly scored report |

## Next Steps

1. Publish SOC 2 Type I report via `/security/soc2` endpoint (Q3)
2. Engage external auditor for SOC 2 Type II (Q4)
3. Automate evidence collection in `pillar-audit.yml` workflow
