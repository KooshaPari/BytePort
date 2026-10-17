# BytePort mature obligations — semantic slice 04: security, state, interfaces, distribution

Status: **PROVISIONAL / denominator expansion, not freeze**  
Date: 2026-09-30.

## Security / authorization

### BP-MO-AUTH-002 — credentials are references with provider scope
Source/build/runtime/publication credentials are stored/retrieved through authorized secret references and bound to provider/account/target scope. Evidence and portfolio output MUST NOT contain secret values.

### BP-MO-AUTH-003 — operation authorization is rechecked at mutation boundary
Planning a deployment does not permanently authorize a later mutation. Runtime/build/publication adapters receive an authorization context valid for the exact Operation/Target.

### BP-MO-NET-001 — control-plane reachability is explicit
Loopback-only, LAN/tailnet or externally reachable control-plane modes are explicit configurations with corresponding authentication requirements. Binding `0.0.0.0` is not equivalent to local-only.

### BP-MO-NET-002 — browser/CORS policy is not network authorization
CORS/origin allowlisting cannot substitute for authenticated network/API authorization.

## State / migration

### BP-MO-STATE-001 — schema preserves identity distinctions
Persistence represents Project, SourceSnapshot, ManifestRevision, operations, artifacts, generations, provider resources and observations without overloading one UUID.

### BP-MO-STATE-002 — additive migration preserves historical ambiguity honestly
Existing Project/deployments-map rows migrate without retrospectively asserting selected-source/artifact correctness they never proved.

### BP-MO-STATE-003 — historical placeholder deployments remain distinguishable
Legacy `alpine:latest`/unknown-artifact deployments cannot become VERIFIED RealizedDeployments solely because they exist in storage.

### BP-MO-STATE-004 — evidence/history is append-preserving
Reconciliation/update/rollback does not erase prior operation/resource/observation history needed to explain product state.

## Common application interfaces

### BP-MO-API-001 — CLI/Desktop/API invoke the same domain operation
Surface-specific handlers translate into the same application-layer Build/Runtime/Publication operations and identities.

### BP-MO-API-002 — planning and execution are distinguishable
A plan-only CLI/UI result cannot be presented as deployed/realized.

### BP-MO-API-003 — operation observation is resumable
A caller can reconnect/reopen and observe an existing durable operation rather than creating another mutation.

### BP-MO-API-004 — errors preserve uncertainty
Timeout/adapter error responses distinguish definite rejection, definite failure, accepted-but-unknown and reconciliation-required outcomes.

## Installation / distribution

### BP-MO-DIST-001 — installed control plane has verifiable artifact identity
Desktop/CLI/server distribution identifies exact build/version/signature provenance appropriate to supported platform.

### BP-MO-DIST-002 — update/rollback preserves domain-state compatibility
Schema/config/credential references are migrated or explicitly blocked; update cannot silently orphan provider resources/operations.

### BP-MO-DIST-003 — uninstall disposition is explicit
Local product state, credentials, build cache/artifacts and remote provider resources have explicit independent retain/remove policies. Uninstall MUST NOT silently destroy remote resources unless specifically authorized.

### BP-MO-PLAT-001 — platform support is journey-specific
Desktop/CLI/API support claims require installed journey evidence, not merely successful frontend/backend compilation.

## Operational quality overlays

Subjects may carry:
- source-resolution latency/freshness;
- build duration/cache effectiveness;
- operation/reconciliation latency;
- provider API retry limits;
- observation freshness;
- publication latency;
- credential exposure risk;
- accessibility/usability;
- local resource overhead.

Targets are not invented here. They require baseline/alternative measurements or explicit product objectives.

## Exit
Requires threat/reachability experiment, schema migration plan/fixture, reconnectable operation API prototype, installed distribution fixtures and explicit remote-resource uninstall policy.
