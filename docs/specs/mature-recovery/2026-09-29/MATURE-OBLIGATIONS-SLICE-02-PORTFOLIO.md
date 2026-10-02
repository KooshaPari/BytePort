# BytePort mature obligations — semantic slice 02: portfolio projection and publication

Status: **PROVISIONAL SEMANTIC SLICE / NOT FULL CONTRACT**
Date: 2026-09-30.
Source snapshot: 0232cca16fedb7963a8c6f556dc5eee5c8c1674e.

This slice derives productization obligations from direct November 2024 user intent, the last live pre-ejection Demonstrator implementation, current credential substrate, and the modern PortfolioProjection/PortfolioPublisher boundary.

## BP-MO-PROJ-001 — portfolio projection references exact product truth

**Statement.** A PortfolioProjection MUST reference the Project, SourceSnapshot, ManifestRevision, RealizedDeployment, BuildArtifact/provider identities as applicable, and the freshness of runtime/endpoint observations used to construct it.

**Counterexample.** a project card says a deployment is live using only an old AccessURL stored at deploy time.

## BP-MO-PROJ-002 — verified and generated fields are authority-separated

**Statement.** Verified facts, user-authored metadata, generated/inferred copy, and manually reviewed overrides MUST remain distinguishable in projection data and evidence.

**Historical lesson.** the old Demonstrator asked an LLM to fill one broad template from project state without an explicit authority partition.

## BP-MO-PROJ-003 — secrets never enter projection context by default

**Statement.** Portfolio generation MUST receive an explicit secret-free allowlist of project/deployment fields. Decrypted AWS, Git, LLM, portfolio, environment, or provider secrets MUST NOT be serialized into prompts or publisher payloads unless an accepted field specifically requires a protected reference.

## BP-MO-TPL-001 — template contract has identity/version

**Statement.** A fetched template/schema/component contract MUST have target/publisher identity and version/content identity sufficient to reproduce or explain the generated projection.

**Counterexample.** portfolio backend changes template shape between generation and retry and BytePort silently mutates the project payload.

## BP-MO-PUB-001 — publication is a distinct operation

**Statement.** Publishing a PortfolioProjection MUST create/reference a PortfolioPublication operation distinct from DeploymentOperation.

**Positive acceptance.** healthy deployment + failed publisher yields deployment REALIZED and publication FAILED/RETRYABLE, not deployment failure.

## BP-MO-PUB-002 — publisher adapter is target-scoped

**Statement.** PortfolioPublisher credentials, endpoint/repository/path and capability contract MUST be bound to the exact publication target.

**Counterexample.** source GitHub repository credentials are reused to write an unrelated portfolio repo without explicit target authorization.

## BP-MO-PUB-003 — API publication is idempotent or reconcilable

**Statement.** Retrying a publication after timeout/lost response MUST not create duplicate project entries. The adapter MUST expose an idempotency/reconciliation mechanism or explicit UNKNOWN state.

## BP-MO-PUB-004 — Git fallback is an adapter, not shell scripting

**Statement.** If the accepted Git/server-side fallback remains enabled, it MUST use explicit publication repository/path/ref identity, authorization, content diff, commit identity and conflict/retry semantics.

**Current mapping.** no product Git publication path exists at the frozen SHA.

## BP-MO-PUB-005 — API and Git fallback preserve one logical publication identity

**Statement.** Switching transport from portfolio API to authorized Git fallback MUST update/reconcile the same logical project publication rather than create divergent duplicate portfolio entries.

## BP-MO-PUB-006 — publisher outage cannot corrupt deployment truth

**Statement.** Portfolio API unavailability, template failure, generation failure or Git conflict MUST NOT roll back or misreport a healthy realized deployment unless the user explicitly requested an atomic deployment+publication policy.

## BP-MO-OBS-002 — stale/dead endpoints are explicit

**Statement.** A projection that includes live endpoint claims MUST carry observation freshness and surface stale/unreachable state rather than silently preserving a verified-live label.

## BP-MO-GEN-001 — generation provider is replaceable and bounded

**Statement.** LLM-assisted text/media enrichment is an adapter over a structured projection schema; failure/unavailability MUST preserve a non-generated baseline projection when the accepted journey allows it.

**Historical lineage.** old Demonstrator supported multiple provider configurations; the mature contract preserves replaceability, not its module layout.

## BP-MO-REVIEW-001 — publication policy can require review

**Statement.** Generated public-facing content MUST support an accepted review/approval policy before external publication where configured; generated output is not self-authorizing.

## BP-MO-PRIV-001 — public projection is explicit allowlist

**Statement.** The public portfolio representation MUST be derived from an allowlisted public schema, not from serializing the internal Project/User/Provider state object.

## Current implementation trace

Credential onboarding, encrypted storage and SSRF-hardened portfolio endpoint validation are present. Projection generation, template fetch for publication, publication POST, Git fallback, public render path and PortfolioPublication state are absent at the frozen SHA. The last live pre-ejection NVMS Demonstrator provides historical behavioral evidence only.

## Slice exit

Promote only after a provider-independent publisher fixture closes API publish/idempotent retry/stale endpoint/redaction, a Git fallback fixture targets a separate authorized repository/path, and the deployment/publication state split is verified end-to-end.
