# BytePort non-code completion ledger — v1.0

Date: 2026-10-01.
Purpose: authoritative denominator for specification, documentation, oracle design, research, architecture and governance work. Runtime/provider implementation proof is separate.

## Finalized non-code foundations

| Family | State | Finality |
|---|---|---|
| direct product thesis | CLOSED | repository + compact declarative desired state → generalized heterogeneous infrastructure/deployment lifecycle |
| mature boundary | CLOSED | selected-app and portfolio are projections; cloud/bare-metal/general infrastructure are core horizon |
| authority model | CLOSED | direct user intent distinguished from ADR/source/assistant suggestion |
| ontology | V1.2 CLOSED FOR CURRENT AUTHORITY | desired/realized resource graph, target, build, operation, generation, observation, reconciliation, destruction represented |
| mature-first/stages | CLOSED | selected-app CVP uses mature identities and can widen without replacement |
| evidence identity | CLOSED | source/manifest/artifact/operation/generation/resource/verifier exactness required |
| MACE/autograder | CLOSED | multidimensional independent grading and hard false-green controls |
| legacy migration doctrine | CLOSED | historical deployments remain legacy_unverified absent exact evidence |
| provider/build abstraction doctrine | CLOSED | capabilities/adapters, not AWS/Docker/NanoVMS product identity |
| orphan/scope doctrine | CLOSED | transport/planner/old Demonstrator/current placeholders classified |
| work-package DAG | CLOSED | machine-readable dependencies/gates/evidence |
| destructive lifecycle doctrine | CLOSED | uninstall != destruction; exact authorized DestructionIntent required |

## Contract families represented

- SourceReference/SourceSnapshot;
- ManifestRevision/schema/version/digest;
- DesiredResourceGraph/dependencies;
- runnable/non-artifact/managed resources;
- BuildEngine/ArtifactResolution/BuildArtifact/provenance;
- Target/PlacementConstraint;
- ResourcePlan;
- RuntimeOperation/ExternalOperationRef;
- DeploymentGeneration;
- RealizedResource/ProviderResource;
- observation/freshness;
- reconciliation create/read/update/replace/delete/no-op/unknown;
- restart/lost-response/idempotency;
- rollout/rollback;
- authorization/network boundary;
- secrets;
- CLI/Desktop/API convergence;
- portfolio projection/publication;
- migration/history;
- install/update/uninstall;
- cloud/bare-metal capability differences;
- evidence/trace/stages/journeys.

## SOTA/alternatives status

### Closed architecture-significant conclusions
- desired resource graph + explicit plan/reconcile separation is validated by OpenTofu-style graph/plan prior art;
- desired resource vs external realized resource/provider controller split is validated by Crossplane-style reconciliation;
- provider plugin/capability architecture is validated by Pulumi/Crossplane prior art;
- bare-metal low-level provisioning should bootstrap existing engines before custom BMC/PXE/imaging;
- standalone Ironic + Bifrost is the leading bare-metal bootstrap candidate;
- Tinkerbell/MAAS remain comparison candidates rather than product identity;
- provider-specific features remain capability extensions;
- BytePort owns cross-target desired state, identity, policy, lifecycle, reconciliation, UX and evidence.

### Empirical questions intentionally open
- Ironic/Bifrost vs Tinkerbell/MAAS operational fit for the user's environment;
- exact BuildEngine support matrix;
- provenance implementation;
- numeric deployment/reconciliation targets;
- exact network-mode defaults;
- portfolio publication destination policy.

## Journey finality

Mature journey families are defined:
- auth/source;
- source→manifest;
- desired graph→artifact/resource plan;
- operation→target realization;
- uncertain mutation→reconcile;
- observe/update/rollback/destroy;
- portfolio projection/publication;
- cross-interface operation;
- install/update/uninstall.

Selected-app remains the first verification slice; generalized mixed-resource reconciliation is the second.

Contracts/oracles are final enough to implement. Runtime closure is separate.

## Test/oracle design finality

Adversarial families represented:
- branch movement;
- invalid/unknown manifest;
- wrong artifact returning 200;
- mutable tag;
- provider response loss;
- duplicate operation/fingerprint conflict;
- delayed visibility;
- wrong resource ID;
- zero/multiple resource selection;
- rollout generations;
- unauthorized principal/target;
- network/CORS confusion;
- publication failure;
- legacy migration false verification;
- desired/observed UNKNOWN;
- destructive reconciliation;
- local uninstall vs remote resources;
- heterogeneous target capability mismatch.

## Remaining non-code blockers

1. **Claude corpus not yet ingested.** User says it contains highly relevant conversations.
2. **Network-mode default/exposure policy.** Loopback vs authorized LAN/tailnet/external default is a product/operational decision still requiring authority and threat review.
3. **Fresh independent falsification review** after Claude delta and remaining empirical gates.
4. **Empirical architecture gates:** B03 execution, build provenance, generalized reconciliation provider fixture, bare-metal engine comparison, portfolio publisher experiment.
5. **Numeric quality targets** require measured baselines/user objectives.
6. **Manifest schema authority beyond recovered minimum.** Minimal recovery-v1 is sufficient for the first slice; richer INFRASTRUCTURE/PORTFOLIO/BUILD semantics remain versioned extensions until authority/evidence closes them.

## Finality rule

Reopen only for new authority/source/prior-art falsification/concrete uncovered behavior or empirical evidence. Additional document volume by itself is not evidence of incompleteness.
