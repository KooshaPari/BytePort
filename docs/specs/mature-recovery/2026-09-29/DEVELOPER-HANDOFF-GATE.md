# BytePort developer-agent handoff gate — v0.2

Status: **PARTIAL / TWO-TIER HANDOFF**  
Date: 2026-09-30.

Developer agents may work only inside the tiers below. BP-F03/BP-F04 are native-reproduced risks; newer duplicate/journal/build prototypes remain pending/non-evidence.

## Tier A — structurally authorized now

### BP-DEV-A1 — explicit domain identities
Introduce additive domain/storage types for:
SourceReference, SourceSnapshot, ManifestRevision, BuildOperation, BuildArtifact, DeploymentIntent, RuntimeOperation, ProviderResource, RealizedDeployment and Observation.

No destructive migration or public behavior replacement yet. Existing Project/User data must remain readable.

### BP-DEV-A2 — BuildEngine contract
Implement the capability interface and test doubles defined in BUILD-ENGINE-ADAPTER-CONTRACT.md.

Allowed initial adapters:
- immutable prebuilt artifact verifier;
- disposable BuildKit/Dockerfile adapter.

Do not make Docker the product identity.

### BP-DEV-A3 — common application operation interface
Create one domain deploy interface consumable by CLI/Desktop/API, initially behind existing surfaces or test harnesses.

It must expose Operation identity rather than hide provider mutation behind route success.

### BP-DEV-A4 — trace/evidence enforcement
Machine-validate exact source/artifact/operation/provider/evidence identities and quarantine legacy/generated completion catalogs.

## Tier B — experiment-dependent / bounded remediation

### BP-DEV-B1 — provider-resource termination
BP-F03 proves current behavior wrong, so remediation design is authorized. Production merge must include:
- zero/one/multiple resource selection semantics;
- ownership/target checks;
- exact provider identity negative control;
- no ProjectID fallback.

### BP-DEV-B2 — durable operation journal
Prototype BP-AD-01 now. Production merge waits for isolated journal-model + duplicate-deploy receipts and restart/reconcile provider fixture.

### BP-DEV-B3 — selected source/manifest path
Prototype SourceReference→SourceSnapshot→ManifestRevision resolution. Production deploy integration waits for the delegated-build selected-app prototype and immutable artifact receipt.

### BP-DEV-B4 — BuildEngine adapters
BuildKit/CNB/prebuilt/external-CI experiments are authorized. First-party support decisions wait for comparative receipts.

### BP-DEV-B5 — runtime reconciliation
Prototype lost-response/restart/compensation semantics. Do not treat compensation as proof the remote operation never happened.

### BP-DEV-B6 — PortfolioPublisher
Prototype independently of deployment acceptance. Publication failure must not rewrite deployment truth.

## Explicitly prohibited

- rebuilding NanoVMS inside BytePort;
- sole Docker-specific product architecture;
- wholesale historical Demonstrator restoration;
- desktop-specific deployment domain;
- deleting CLI-first authority without direct supersession;
- provider/cloud breadth before one vertical slice closes;
- route-200/document-count/CI-green completion claims.

## First vertical slice promotion gate

One fixture must independently prove:

`SourceReference → SourceSnapshot → ManifestRevision → BuildOperation → immutable BuildArtifact → RuntimeOperation → ProviderResource → application Observation → RealizedDeployment`

and exercise retry, restart, exact stop, wrong-artifact rejection and one portfolio projection.

## Verdict

**Tier A developer handoff: READY.**  
**Tier B prototype/remediation handoff: READY with per-package merge gates.**  
**Whole mature BytePort implementation: NOT READY.**
