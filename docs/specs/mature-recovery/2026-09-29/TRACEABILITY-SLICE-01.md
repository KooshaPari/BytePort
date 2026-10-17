# BytePort bidirectional trace slice 01 — source, build, lifecycle, publication

Status: **DRAFT STRUCTURAL TRACE / not full graph**  
Date: 2026-09-30.  
Frozen source: `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`.

## Trace nodes

| Trace ID | Product obligation / decision | Authority / provenance | Implementation relation | Oracle/prototype | Exact evidence | Current state |
|---|---|---|---|---|---|---|
| BP-T01 | Mutable source ref resolves once to immutable SourceSnapshot | direct user intent + PRD/FR + BP-MO-SRC-001 | live web sends repository object/id; mounted handler does not resolve ref/commit | branch-move fixture | not implemented | **missing** |
| BP-T02 | Exact ManifestRevision validated/bound before deployment | direct user intent + accepted manifest ADR/FR + BP-MO-MAN-001 | mounted deploy path does not accept/parse manifest | invalid/changed manifest controls | current source mapping | **missing** |
| BP-T03 | BuildArtifact has immutable identity/provenance | BP-MO-ART/BUILD + BP-AD-02; external BuildKit/CNB research | mounted handler uses mutable unrelated `alpine:latest` | selected-app delegated-build prototype | prototype run currently pending | **current implementation conflicts** |
| BP-T04 | ProjectID != ProviderResourceID; stop exact realized resources | BP-MO-ID/STOP + ontology | deploy stores sandboxResp.ID; terminate uses project.UUID | distinct-ID handler test | candidate `934b0456...`, job `109572633061` | **NATIVE COUNTEREXAMPLE REPRODUCED** |
| BP-T05 | Durable product Operation precedes uncertain remote side effects | BP-MO-OP-001 + BP-AD-01 | current provider deploy precedes Project persistence; no operation journal in inspected path | persistence-window probe | candidate `aae3fbaa...`, run `36686241335`, job `109792659013` | **NATIVE RISK REPRODUCED** |
| BP-T06 | Unknown/reconciling state survives restart/lost response | BP-MO-OP-002 + BP-AD-01 | no mapped operation state machine | executable journal reference model + future provider fixture | model committed; isolated workflow pending | **architecture-model pending** |
| BP-T07 | Same operation+fingerprint is retry; same ID+different fingerprint rejects | BP-AD-01 | current live API accepts no OperationID | reference model + duplicate-deploy observational probe | duplicate probe exact isolated workflow pending | **implementation missing** |
| BP-T08 | Desired vs observed state distinct; selected application independently observable | BP-MO-OBS/APP | control response currently sufficient for handler success; no source/artifact live identity | `/__byteport_probe` fixture | build prototype pending | **missing** |
| BP-T09 | Build engine is replaceable and evidence-bearing | BP-AD-02 + BuildEngine contract | Tauri plan names placeholder `artifact build`; no common application layer | Docker/BuildKit/CNB prototype ladder | P0 Docker workflow pending | **architecture candidate** |
| BP-T10 | CLI/Desktop/API project onto one domain Operation | accepted ADR-005 + BP-AD-03 | Tauri deploy is plan-only; web/backend performs unrelated real mutation | same-intent cross-projection negative controls | PASS-7 source mapping | **architectural drift** |
| BP-T11 | PortfolioProjection derives from verified source/deployment facts | direct Nov 2024 user intent + historical implementation + BP-MO-PORT | current named J3 implementation absent; credentials remnants exist | stale/wrong/generated/secret controls | historical pre-ejection corroboration | **mature obligation / current gap** |
| BP-T12 | PortfolioPublisher separate from deployment result; API + authorized Git fallback | direct user intent + BP-MO-PUB | historical API publisher existed; current product publication path not mapped | independent publish failure/retry/idempotency fixture | historical source evidence only | **current implementation missing** |

## Reverse trace examples

### Wrong stop target

`TerminateInstance → project.UUID stop query → BP-T04 → distinct-ID test → native failure → BP-AD-01/Provider lifecycle contract`.

### Placeholder deployment

`web deployProject → POST /deploy → deployRequest repository metadata → alpine:latest SandboxConfig → BP-T01/T02/T03/T08 → selected-app prototype → BP-AD-02`.

### Portfolio

`direct user deployment→portfolio intent → historical Demonstrator template/API path → BP-T11/T12 → PortfolioProjection/Publisher architecture → current implementation gap`.

## Orphan / accidental-architecture candidates

- generic `byteport-transport`/codec/UI ports from June agent focus work must trace to BP-T01–12 or another accepted mature obligation before they count as product scope.
- Tauri deploy DAG planner is a scaffold/proposal until it creates/observes the same domain Operation as web/API.
- historical NanoVMS Demonstrator code informs portfolio behavior but is not current ownership.
- root README desktop-only positioning is non-authoritative agent-generated documentation and cannot trace to mature scope.

## Slice exit

Close after P0/P1 build prototype, duplicate-operation and journal-model receipts, selected-app runtime fixture, provider reconciliation prototype, portfolio publisher adapter experiment, and accepted authority/architecture review.