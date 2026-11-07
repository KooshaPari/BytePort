# BytePort selected-app vertical-slice contract — v0.1

Status: **DESIGN/ORACLE CONTRACT / Tier A model work + Tier B lifecycle integration**  
Date: 2026-09-30.

## Current structural blocker

Frozen `Project` persists:
- project/user metadata;
- mutable Repository metadata;
- `DeploymentsJSON` map of provider instances;
- readme/description/platform/access URL/type.

It does not persist:
- SourceSnapshot;
- ManifestRevision;
- Component/ArtifactResolution;
- BuildOperation/BuildArtifact;
- product RuntimeOperation;
- DeploymentGeneration;
- ProviderResource as a typed durable entity;
- Observation;
- RealizedDeployment evidence.

Therefore the selected-app slice requires additive domain/storage identities. Reinterpreting existing Project/deployments JSON would manufacture historical certainty.

## Additive storage migration

Introduce new tables/records conceptually:

- `source_snapshots`
- `manifest_revisions`
- `components`
- `build_operations`
- `build_artifacts`
- `deployment_intents`
- `runtime_operations`
- `external_operation_refs`
- `deployment_generations`
- `provider_resources`
- `observations`
- `portfolio_projections`
- `evidence_receipts`

Exact physical schema is implementation work; these identities are the semantic minimum.

Legacy Project rows:
- remain readable;
- may reference imported/legacy provider instances;
- are classified `artifact_identity=unknown`, `source_snapshot=unknown`, `verification=legacy_unverified` unless exact historical evidence exists;
- MUST NOT be silently promoted to RealizedDeployment.

## Vertical slice fixture

One disposable Git fixture with:
- exact requested ref;
- manifest containing one runnable component;
- probe endpoint returning source commit + manifest digest + artifact identity marker;
- independent unrelated runtime fixture used as stop-safety control.

## Oracle bundle

### BP-VS-01 source resolution
Mutable requested branch/ref resolves once to exact commit. Moving branch afterward does not alter SourceSnapshot.

### BP-VS-02 manifest identity
Manifest bytes are read at SourceSnapshot, schema-validated and content-digested. Invalid schema stops before build/runtime mutation.

### BP-VS-03 artifact resolution
Component resolves through selected BuildEngine or verified prebuilt artifact. Success requires immutable artifact identity.

### BP-VS-04 provenance
Build receipt binds SourceSnapshot, ManifestRevision, BuildEngine/version and BuildArtifact.

### BP-VS-05 durable runtime operation
RuntimeOperation + request fingerprint persist before provider mutation.

### BP-VS-06 provider resource identity
Provider-assigned resource identity persists separately from Project/Operation.

### BP-VS-07 selected application observation
Independent probe proves running application corresponds to requested source/manifest/artifact; provider HTTP success alone fails acceptance.

### BP-VS-08 control-plane restart
Restart BytePort after provider success. Reopen durable operation/generation/resource and continue observation without creating a second resource.

### BP-VS-09 lost response
Provider accepts mutation but response is lost. State becomes UNKNOWN/RECONCILING; retry does not blindly duplicate.

### BP-VS-10 duplicate operation
Same OperationID + same fingerprint resumes/observes same intent. Same OperationID + changed fingerprint rejects.

### BP-VS-11 delayed provider visibility
Temporary not-found after uncertain mutation remains UNKNOWN while provider contract permits delay; it is not immediately marked absent.

### BP-VS-12 exact stop
Stop targets ProviderResource identity. Unrelated resource survives. No ProjectID fallback.

### BP-VS-13 rollout generation
Replacement deployment creates candidate generation; old/current/candidate roles remain explicit until observation selects current.

### BP-VS-14 authorization
Wrong Principal/account/target cannot mutate or observe protected operation/resource beyond authorized policy.

### BP-VS-15 session expiry
Expired session token is rejected at mutation boundary. Current native oracle is pending and may falsify/confirm the static concern.

### BP-VS-16 portfolio projection
Realized projection derives verified facts from exact evidence; generated/manual fields remain distinct.

### BP-VS-17 publication independence
Publisher failure/retry does not change RealizedDeployment truth. Same publication operation is idempotent.

### BP-VS-18 local network mode
Desktop/local mode binds according to accepted network policy; CORS is not used as authentication. Remote mode, if supported, requires explicit configuration.

### BP-VS-19 CLI/Desktop/API convergence
Each projection creates/observes the same domain Operation semantics; plan output is never deployment success.

### BP-VS-20 uninstall
Removing local BytePort leaves remote resource untouched unless explicit resource-destruction authorization is given.

## Evidence receipt

The final slice receipt must bind:
- source/manifest/build/runtime adapter revisions;
- BytePort candidate;
- dependency lock;
- principal/target;
- operation/generation/resource identities;
- raw build/provider/probe/publication artifacts;
- timestamps/freshness;
- every negative-control result.

## Promotion gate

No “vertical slice closed” claim until BP-VS-01..20 are either executed for the slice or explicitly classified non-applicable with accepted rationale.

Unit-only models do not close public-path criteria.
