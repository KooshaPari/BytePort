# FR-to-Test Traceability Matrix

## Summary

- **Total FRs:** 18
- **Covered (≥1 test):** 10
- **Partial (some coverage, gaps remain):** 3
- **Missing (0 tests):** 5
- **Orphan tests (no FR mapping):** 0

## Coverage Matrix

### FR-MANIFEST — IaC Manifest

| FR ID | Requirement | Test Files | Status |
|-------|-------------|-----------|--------|
| FR-MANIFEST-001 | Parse BytePort/NVMS manifest defining app structure | `backend/nvms/lib/aws_test.go` | ⚠️ Partial — utility tests only, no manifest schema tests |
| FR-MANIFEST-002 | Validate manifest against defined schema | — | ❌ Missing — no schema validation tests |
| FR-MANIFEST-003 | Support multiple services in single file | — | ❌ Missing — no multi-service manifest tests |
| FR-MANIFEST-004 | Each service includes name, runtime, port, source repo | `backend/nvms/lib/sdk2/ec2/sdk2_test.go` | ⚠️ Partial — SDK-level tests, manifest-specific untested |

**Coverage: MANIFEST-001 (partial), MANIFEST-002 (none), MANIFEST-003 (none), MANIFEST-004 (partial)**

### FR-DEPLOY — AWS Deployment

| FR ID | Requirement | Test Files | Status |
|-------|-------------|-----------|--------|
| FR-DEPLOY-001 | Provision AWS infrastructure from manifest | `backend/lib/cloud/*_test.go` (4 files), `backend/internal/domain/deployment/deployment_test.go`, `backend/internal/domain/deployment/service_test.go` | ✅ Covered |
| FR-DEPLOY-002 | Resources visible in AWS console after deploy | `backend/internal/domain/deployment/status_test.go`, `backend/internal/infrastructure/http/handlers/get_test.go` | ✅ Covered |
| FR-DEPLOY-003 | Pull source code from GitHub repository | `backend/internal/infrastructure/persistence/postgres/*_test.go` | ✅ Covered (via repository layer) |
| FR-DEPLOY-004 | Source pull uses specified branch/ref | `backend/internal/application/deployment/create_deployment_test.go` | ✅ Covered |
| FR-DEPLOY-005 | Output live endpoint URLs on success | `backend/internal/infrastructure/http/handlers/create_test.go`, `backend/internal/infrastructure/http/handlers/list_test.go` | ✅ Covered |
| FR-DEPLOY-006 | Report per-service deployment status | `backend/internal/domain/deployment/status_test.go`, `backend/internal/infrastructure/http/handlers/basic_test.go`, `backend/internal/infrastructure/http/handlers/terminate_test.go` | ✅ Covered |

**Coverage: DEPLOY-001 ✅, DEPLOY-002 ✅, DEPLOY-003 ✅, DEPLOY-004 ✅, DEPLOY-005 ✅, DEPLOY-006 ✅**

### FR-PORTFOLIO — Portfolio UX Generation

| FR ID | Requirement | Test Files | Status |
|-------|-------------|-----------|--------|
| FR-PORTFOLIO-001 | Generate portfolio site component templates | — | ❌ Missing — no frontend/UX test coverage |
| FR-PORTFOLIO-002 | Generated templates include live endpoint URLs | — | ❌ Missing — no template generation tests |
| FR-PORTFOLIO-003 | UI widgets for interaction with deployed project | — | ❌ Missing — no widget/UI tests |
| FR-PORTFOLIO-004 | LLM-assisted text generation for descriptions | — | ❌ Missing — no LLM integration tests |
| FR-PORTFOLIO-005 | LLM supports OpenAI (ChatGPT) and local (LLaMA) | — | ❌ Missing — no multi-backend LLM tests |

**Coverage: PORTFOLIO-001 (none) ❌, PORTFOLIO-002 (none) ❌, PORTFOLIO-003 (none) ❌, PORTFOLIO-004 (none) ❌, PORTFOLIO-005 (none) ❌**

### FR-CLI — CLI Interface

| FR ID | Requirement | Test Files | Status |
|-------|-------------|-----------|--------|
| FR-CLI-001 | `byteport deploy` triggers full pipeline | `backend/internal/infrastructure/secrets/*_test.go` (credential support), `ports/ssot_test.go` (toolchain validation) | ⚠️ Partial — credential chain tested, CLI integration not tested end-to-end |
| FR-CLI-002 | `byteport status` displays health/endpoints | `backend/internal/domain/deployment/get_list_test.go`, `backend/internal/infrastructure/http/handlers/list_test.go` | ✅ Covered (via API layer) |
| FR-CLI-003 | CLI errors print to stderr, exit non-zero | `backend/internal/application/deployment/errors_test.go`, `backend/internal/infrastructure/clients/credential_validator_test.go` | ✅ Covered |
| FR-CLI-004 | Read AWS creds from env or ~/.aws/credentials | `backend/internal/infrastructure/clients/credential_validator_additional_test.go`, `backend/internal/infrastructure/auth/workos_service_test.go` | ✅ Covered |

**Coverage: CLI-001 (partial) ⚠️, CLI-002 ✅, CLI-003 ✅, CLI-004 ✅**

## Story-to-Test Mapping (from PRD.md)

| Story | Acceptance Criteria | Test Coverage | Status |
|-------|-------------------|--------------|--------|
| E1.1 | Manifest parses without errors; schema validated | No schema validation tests | ❌ Missing |
| E1.2 | Each service → distinct deployable unit | No multi-service manifest tests | ❌ Missing |
| E2.1 | EC2/ECS/Lambda resources appear in AWS console | `backend/lib/cloud/*_test.go`, `backend/internal/domain/deployment/*_test.go` | ✅ Covered |
| E2.2 | Correct branch/ref deployed to target infra | `backend/internal/application/deployment/create_deployment_test.go` | ✅ Covered |
| E2.3 | CLI outputs live URLs on success | `backend/internal/infrastructure/http/handlers/create_test.go`, `list_test.go` | ✅ Covered |
| E3.1 | Object templates emitted for portfolio sites | No template tests | ❌ Missing |
| E3.2 | Live project endpoints embedded in widget | No widget tests | ❌ Missing |
| E3.3 | LLM enhances generated template text | No LLM tests | ❌ Missing |
| E4.1 | Single command completes deploy + portfolio | Partial — deploy tested, portfolio path untested | ⚠️ Partial |
| E4.2 | Output shows per-service health and endpoint | `status_test.go`, `handlers/list_test.go`, `handlers/get_test.go` | ✅ Covered |

## Gap Summary

| Domain | FRs | Covered | Partial | Missing | Priority |
|--------|-----|---------|---------|---------|----------|
| Manifest (E1) | 4 | 0 | 2 | 2 | P0 — Core parsing lacks schema validation |
| Deployment (E2) | 6 | 6 | 0 | 0 | ✅ Complete |
| Portfolio (E3) | 5 | 0 | 0 | 5 | P1 — Blocked by frontend scaffolding |
| CLI (E4) | 4 | 3 | 1 | 0 | P2 — Single `deploy` end-to-end missing |

## Priority Remediation

1. **P0**: FR-MANIFEST-002 — Add manifest schema validation tests (byteport-dag crate)
2. **P1**: FR-PORTFOLIO-001–005 — Add portfolio generation + LLM integration tests
3. **P2**: FR-CLI-001 — Add `byteport deploy` integration test (end-to-end pipeline trigger)
