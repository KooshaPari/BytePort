# BytePort — source coverage ledger, pass 1

Status: **OPEN DENOMINATOR / BLOCKING AUTHORITY CONTRADICTION**. Observation date: 2026-09-29. Product source: `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`; registry: `85d7cd00cf59c379c05b740e8130a85b0d5bd31b`. See `SNAPSHOT.json`. This additive recovery record does not silently replace SPEC.md or accept registry prose merely because it is newer.

A file-open receipt is not semantic closure. Each source needs meaning, authority, extracted obligations/non-normative disposition, contradiction resolution, reachable implementation mapping, stage/journey implications and oracle consequences. Full family enumeration is still open; no denominator-based completion percentage is asserted.

## Authority and identity

Keep USER_INTENT, ACCEPTED_DESIGN, CURRENT_IMPLEMENTATION, HISTORICAL_IMPLEMENTATION, PROPOSAL, EXPERIMENT, ASSISTANT_SUGGESTION and EXTERNAL_PRIOR_ART separate. Conversation retrieval summaries preserve author/date but are not a substitute for original transcript exports. Registry intent labels and bound-prompt counts are imported assertions until provenance resolves. NanoVMS is a separately branded runtime/dependency, not an alias for BytePort.

## Inspected sources

All product files are pinned to the product SHA above; registry files to the registry SHA.

| ID | Source / extent / blob where available | Meaning and classification | Consequence for obligations, journeys and verification | Resolution |
|---|---|---|---|---|
| BP-S01 | `README.md`, full; `92fd66da6947fa0039fa10b3e65078bcb98ff594` | Supporting local desktop/container description, not acceptance | Desktop packaging and local target claims; module-count contradiction within README. Own-ecosystem uniqueness is not external differentiation. | PARTIAL/CONTRADICTORY |
| BP-S02 | `SPEC.md`, lines 1–190; `7d4f4160b0cf30dcf9cfa5bdd97cf20fb24ffd43` | Canonical-labeled May design: manifest-driven AWS deployment plus portfolio generation | Mature source→manifest→runtime→published-project journey; conflicts with local-only narrowing. Historical crypto/architecture statements require source validation. Remaining sections not read. | PARTIAL/CONTRADICTORY |
| BP-S03 | Registry `docs/intent/BytePort.md`, full; `823c09ff306b015092c069e15bc814c08873cd3b` | Imported intent interpretation, last_verified 2026-09-20 | Lists 39 bound prompt references but explicitly says curated corpus absent/unresolved. Cannot infer 39 reviewed user statements. Desktop-only mature scope lacks recovered authorized pivot. | BLOCKING PROVENANCE/SCOPE CONFLICT |
| BP-S04 | `backend/README.md`, full; `071ba31fa8c767076b1127082b40fac32b899788` | Supporting retirement note | Says old github.com/byteport/api and bytebridge Fermyon modules retired 2026-09-20; root README and registry still describe pending removal/two backends. Verify deletion history, preserve historical record. | PARTIAL |
| BP-S05 | `backend/byteport/main.go`, full; `552a8f78d73324c90143be151add6ce7a2be3cd6` | CURRENT_IMPLEMENTATION: mounted Gin routes and startup | /deploy and /terminate wired under AuthMiddleware. /health and /metrics public. Server binds 0.0.0.0, not loopback-only; CORS is not a substitute for a bind/access policy. Registry healthz/presign claims do not match this router. | PARTIAL; runtime network/auth tests needed |
| BP-S06 | `backend/byteport/routes/deployment.go`, full through tail; `cc9c18f006dc0680750884e456b90af3b0f67246` | CURRENT_IMPLEMENTATION: deploy/terminate | Selected project/repository becomes hard-coded alpine:latest/native upstream request. Distinct returned sandboxResp.ID stored in deployment; terminate sends project.UUID instead. Remote success precedes database persistence, with no compensation visible here. Independent identity, payload and crash-window controls required. | PARTIAL; native upstream/database tests outstanding |
| BP-S07 | Frontend `src/lib` directory listing, not implementation bodies | Implementation inventory only | api.ts/config.ts/git.ts/components found; existence is not proof of UI→request reachability. UI and packaged Tauri journeys not yet closed. | OPEN |
| BP-S08 | Searches in CHARTER.md, PLAN.md, API_REFERENCE.md and session route/auth/coverage records | Supporting/historical snippets | NVMS/odin.nvms/portfolio lifecycle; prior route observations are dated, not current greens. Full source reads and evidence identities required. | PARTIAL |
| BP-S09 | Registry ByteBridge alias search in docs/cvp, boundary, intent | Imported boundary prose / stale implementation claims | Pending bytebridge removal conflicts with backend retirement note; healthz/presign claims conflict with mounted router. Not evidence of a callable presign product route. | PARTIAL/CONTRADICTORY |
| BP-S10 | Conversation retrieval: 2024-11-20, 2024-11-30, 2024-12-19 and 2026-08-25 | USER_INTENT as attributed by retrieval; raw transcript not materialized | Git/AWS/NVMS deployment, project/instance hierarchy and portfolio productization recovered independently of current name-only search. Separate runtime branding explicitly stated. No accepted removal of this horizon found in inspected context. | PARTIAL; authority reconciliation remains blocking |

## Source-family denominator

| Family | Known surfaces / inventory work | Disposition and required closure |
|---|---|---|
| Product intent | BP-S01–03,10 | CONFLICT; recover pivot history, not date-based supersession |
| Alias/concept lineage | BytePort/byteport; ByteBridge experiment; .nvms, nanoFile, odin.nvms; Git-to-cloud, portfolio/Slick | PARTIAL; search each independently; NanoVMS is dependency, not product alias |
| README / docs | README, backend README, architecture/install/deploy documentation | PARTIAL; current-vs-retired and install claims |
| ADRs / decisions | Registry ADR-ECO-020 and local reviews pointers | OPEN; full decisions and accepting authority |
| Specifications | SPEC, CHARTER, PLAN, journey docs and DAG requirements | PARTIAL; preserve existing IDs, semantically reconcile all accepted obligations |
| Backend source | backend/byteport lib/routes/models | PARTIAL; handlers and side effects beyond deploy not fully mapped |
| Rust source and machine interfaces | transport, CLI, DAG, telemetry and desktop shell | OPEN; exports/imports and actual consumers |
| API/router | BP-S05–06, authentication and callback surfaces | PARTIAL; mounted route→handler→storage/runtime trace |
| CLI | byteport-cli, deployment/start/task scripts | OPEN; installed interface and actual command effects |
| MCP | Registry assigns elsewhere | OPEN; validate boundary, do not invent first-party scope |
| UI/screens/components | SvelteKit routes/components/api/config and Tauri invoke handlers | OPEN; true user path, errors and state restoration |
| State/schema/migrations | models.Project/Instance/users/data; SQLite/GORM | PARTIAL; provider ID vs product ID, operation journal, crash and migration semantics |
| Tests/oracles | smoke, route/session fixtures, native desktop tests | OPEN; assert requested artifact and distinct provider IDs, not only HTTP 200 |
| CI/build | Rust/Go/frontend matrices and reusable workflow dependencies | OPEN; current checks, skipped/empty/error semantics and toolchain pins |
| Deployment/distribution | Tauri packaging/sidecar startup, install scripts/signing | OPEN; clean installed target and dependency boot |
| Release/lifecycle | release workflows, stable promotion, update/rollback/uninstall | OPEN; acceptance and user-data preservation |
| Security | bind/CORS/auth, credentials/keyring/encryption, GitHub OAuth, Tauri CSP | PARTIAL; local control plane trust, actor/resource ownership, exposure and credential flows |
| Observability/reliability | metrics/health/OTel, provisioning and restart paths | PARTIAL; health != selected application readiness; fail-after-side-effect recovery |
| Integrations | NanoVMS, AWS, GitHub, portfolio/LLM, shared Rust crates | OPEN; manifests/locks and frozen dependency contracts; no third product program |
| Historical implementations | ByteBridge/Fermyon, old Go API, Rust/Loco references | PARTIAL; full DAG and retirement/rename evidence |
| Registry records/mirrors | intent/boundary/CVP/project/atlas/handbook records | PARTIAL; duplicate and stale claims require single authority-aware index |
| External standards/research | Compose/OCI/IaC/desktop/incremental state and partial failure | PARTIAL; current versions, licenses, health and experiments still needed |
| Quality/support/accessibility | desktop usability, documentation and platform behavior | OPEN; shared overlays tied to exact stages and configurations |
| Auxiliary/generated/assets | marketing/audit/catalog/codegen assets | OPEN; explicit non-normative classification; no generic catalog admitted to grading |

## Initial contradiction ledger

- **BP-C01 (blocking):** recovered user Git-to-cloud/productization horizon and SPEC.md vs registry mature desktop/local-only narrowing. Preserve broad horizon provisionally; a local-first stage does not delete mature cloud/portfolio obligations. Find the actual accepted pivot, or reconcile the registry interpretation.
- **BP-C02:** root README's current two-backend claim vs backend retirement note. Source/diff verification outstanding; no cleanup deletion authorized by this dossier.
- **BP-C03:** registry presign/healthz backend description vs actual mounted /health/no presign router. A reusable library is not an HTTP route.
- **BP-C04:** intended selected-project deployment vs hard-coded alpine sandbox. Separate successful sandbox allocation from successful application deployment.
- **BP-C05:** product project identity vs provider sandbox identity in termination. Test nonidentical IDs and exact remote effects.

## Evidence boundary

No deployment to cloud or paid resource was performed. No native desktop, Go application, Rust application or integration suite was executed. Local clone failed DNS resolution. The source evidence establishes code paths and contradictions, not their remediation or a passing installed product. The source ledger, mature-contract mapping, research and independent adversarial review remain incomplete.
