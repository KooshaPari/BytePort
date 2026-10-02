# BytePort — recovered mature horizon and vertical-slice contract, pass 1

**DRAFT / AUTHORITY RECONCILIATION REQUIRED.** Inspected source `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`, 2026-09-29. This record does not automatically supersede SPEC.md, existing requirement IDs, or authorized decisions. See SOURCE-COVERAGE-LEDGER, SEMANTIC-FINDINGS and the PhenoRegistry archaeology/SOTA dossiers.

## Product horizon, not an implementation-shaped definition

Recovered user-attributed November/December 2024 and August 2026 statements, together with SPEC.md's opening, describe **self-service deployment and productization**: selected source plus manifest becomes a deployed application on an owned target, with lifecycle control, observability and portfolio/project presentation. NanoVMS is a separate runtime/dependency, not BytePort's product identity.

September registry prose instead narrows the product to local-only desktop deployment while its underlying prompt corpus is unresolved. A targeted follow-up conversation search found September portfolio/governance work and desktop packaging discussion but no user-accepted removal of cloud deployment or portfolio generation. This is absence within inspected evidence, not exhaustive proof that no pivot occurred.

Working interpretation: local-first desktop is a possible stage/control-plane projection, not permission to erase the broader horizon. Exact supported targets, mobile scope, portfolio publication policy and whether the .nvms DSL remains canonical require accepted decisions. Do not invent a new general cloud control plane or retain one merely because code exists.

## Ontology and invariants under review

Independent projections are needed: capability/journey; product and runtime identity; deployment lifecycle; UI/API/CLI; provider adapters; security/quality; distribution and evidence.

Core candidate entities: Actor/Principal, Project, SourceSnapshot, ManifestRevision, BuildArtifact (content digest), Target and TargetCapability, DeploymentIntent, OperationAttempt, RealizedDeployment, ProviderResourceIdentity, Observation, PortfolioProjection and Release. Existing Project/Instance definitions must be mapped and migrated, not silently reinterpreted.

Critical distinctions:

- Project identity != deployment identity != provider sandbox/resource identity. A project can outlive replacement deployments.
- Accepted requested source != built artifact != allocated sandbox != healthy selected application. Each transition needs its own evidence.
- Desired configuration != last recorded state != fresh provider observation.
- Retrying an operation != creating another deployment. Idempotency must include provider scope and parameter identity.
- Generated portfolio content is a projection with source/deployment provenance, not permission to publish secrets or stale claims.

A provisional lifecycle is requested → validated/planned → authorized → provisioning → observed-ready, with failed/unknown/reconciling/stopping/stopped states. Provider-specific details enrich this model; they must not change identity or destroy durable history. Lifecycle names require alignment with existing code and recovered intent.

## Semantically derived obligation candidates

These are not an accepted exhaustive catalog and do not replace existing IDs. Their sources and counterexamples make the next review concrete.

| Obligation candidate | Source authority / relation | Positive acceptance | Counterexample acceptance |
|---|---|---|---|
| Deploy the selected source/configuration | Recovered user horizon; SPEC; BP-S02/10 | Independent application marker and image/build digest match the selected commit/manifest | Returning 200 for an unrelated alpine image must fail |
| Retain product-to-provider identity mapping | User project/instance hierarchy; BP-S06/11/12 code | Returned provider ID is persisted and used for later observations/stop | Provider ID deliberately differs from project UUID; stop must target the provider ID |
| Recover uncertain side effects | Deployment lifecycle intent; BP-F04; external prior art informs mechanism only | Crash after provider acceptance can reconcile/retry without orphaning or duplicating the operation | Lost response, failed local commit, duplicate/reordered callback and worker replacement cannot produce false success |
| Enforce actor/target scope | Existing mounted authenticated deployment path | Authorized owner acts on its declared target and resource | Wrong user, target, credentials or forged body owner causes no remote side effect |
| Keep presentation tied to evidence | Recovered portfolio/productization intent | Generated preview references current authorized project/deployment facts | Stale or unrelated deployment, unverified generated claim and embedded secrets cannot be published as verified facts |
| Make installed control-plane boundaries explicit | Desktop design records plus actual 0.0.0.0 listener | Supported local/remote access mode is explicitly configured and verified | CORS allowlisting alone cannot serve as proof of local-only reachability |

Each eventual accepted requirement needs stable ID, rationale, source authority, parent capability, dependencies, stage/journey, concrete positive/negative cases, quality overlays, implementation surfaces, verifier strategy and growth disposition. No target count, generic-dimension multiplication, or automatic promotion of these candidates is allowed.

## First-class journeys

| Journey | Actor-to-outcome closure | Current evidence status |
|---|---|---|
| BP-J01 Developer → select source/manifest → inspect plan → authorize → build/deploy → observe own application | The selected application, not a placeholder, is running on the selected target with linked evidence | BLOCKED by source/payload findings; full UI path unverified |
| BP-J02 Operator → reopen/restart during uncertain deployment → reconcile | Durable intent and exact provider identity recover known/unknown state honestly | No native crash-window/restart evidence |
| BP-J03 Developer → update/rollback configuration → observe chosen version | Previous and new deployments retain identity/history and explicit outcome | Mature semantics and provider behavior not yet resolved |
| BP-J04 Developer → generate project/portfolio presentation → review/use or publish | Projection derives from the authorized project's real current state | Intent recovered; current implementation and publication boundary unverified |
| BP-J05 Owner → terminate → verify runtime state → retain appropriate history | Correct owned runtime stopped; unrelated resources survive | Source-level wrong-ID path; native adversarial test outstanding |
| BP-J06 User → install/update/uninstall → preserve or explicitly remove state | Supported signed distribution, credential handling and recovery | Not verified |

## Mature-first stage projections

**CVP proposal:** one supported source/build shape and one target adapter, real BP-J01/BP-J02/BP-J05 closure, plus a minimal evidence-derived portfolio preview for BP-J04. Keep Project/Source/Manifest/Artifact/Target/Operation/ProviderID distinctions from day one. Other target adapters can be explicit unsupported stubs; they cannot return fake deploy success.

**MVP/Beta proposals:** widen source/build/target adapters and update/rollback/recovery capabilities using the same model. Local and owned-cloud adapters must not require changing core identity. **GA:** independently verified supported journey/target/platform matrix, safe migration, signing/update and operational documentation. **Mature:** the complete recovered and accepted horizon, not a feature list invented after an MVP. No stage is assigned to the present implementation from source counts alone.

## Vertical-slice adversarial fixture

Use a disposable real test application whose response embeds the chosen source digest. The runtime returns a resource ID deliberately different from the project ID. Persist a durable operation before remote side effects; implement provider-appropriate idempotency/reconciliation, not an assumption of atomicity between SQLite and an external runtime. Exercise actual persistence, mounted machine interface and actual desktop path, then kill/restart between remote acceptance and local receipt. Verify selected artifact, own target, recovered observation and exact stop target. Generate a local portfolio preview from this evidence; no real public publication or paid cloud deployment is implied by this fixture.

Provider execution engines and manifests should be integrated/adapted where viable; the SOTA dossier compares OpenTofu/Pulumi/Compose/Podman and existing runtime adapters. These are research candidates, not frozen technology substitutions.

## Transition debt and bounded work packages

BP-TD01 existing Project/ID/UUID/Instance mappings need migration and ambiguity handling. BP-TD02 old rows may represent placeholder sandboxes, not selected-source deployments; never retrospectively mark them correct. BP-TD03 retired backend/API/DSL consumers need compatibility disposition. BP-TD04 local credentials, provider identity and operation journal must survive target expansion. BP-TD05 stale portfolio claims need invalidation/provenance, not blind regeneration.

| Package | Dependencies / trace | Exit evidence | State |
|---|---|---|---|
| BP-W1 Source/alias snapshot | BP-S01–12 | Exact source ledger and recovered provenance | Initial pass recorded |
| BP-W2 Authority and lineage reconciliation | W1, original prompts and accepted decisions | Cloud/portfolio/local/desktop and predecessor boundaries resolved | BLOCKING |
| BP-W3 Alternative stack and risky experiments | W2, BP-F01–06, SOTA | Exact artifact/provider-ID/crash recovery comparison | OPEN |
| BP-W4 Ontology and full mature obligations | W2/W3 | Accepted semantic contract and justified stage projections | DRAFT |
| BP-W5 Oracle/trace/reachable mapping | W4 and immutable implementation candidate | Independent public-path evidence and negative controls | DESIGN STARTED |
| BP-W6 Fresh adversarial completeness review | All above | Attempted falsification, no unexplained blocking case | NOT RUN |

No completion percentage or architecture freeze is authorized by this draft.
