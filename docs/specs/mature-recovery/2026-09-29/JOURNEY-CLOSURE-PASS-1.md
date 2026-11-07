# BytePort journey closure matrix — pass 1

Status: **DRAFT / implementation-reachability assessment**  
Date: 2026-09-30.

## BP-J01 — authenticate → select project/source

**Outcome:** authenticated operator identifies a project and a source reference with explicit account/repository authority.

Path:
`login/session → Principal → Project → repository discovery → SourceReference`

Current:
- login/signup/auth middleware mounted;
- project/repository discovery mounted;
- session-expiration semantics under native oracle;
- SourceReference is repository metadata only; no immutable ref resolution.

Classification: **NARROW PARTIAL PATH / NOT CLOSED**.

## BP-J02 — source → validated deployment intent

**Outcome:** requested ref resolves once to SourceSnapshot; exact ManifestRevision validates into component graph/ArtifactResolution/Target.

Current:
- no mounted SourceSnapshot resolver in deploy path;
- no mounted manifest validation in deploy path;
- Tauri deploy planner does not execute this domain transition.

Classification: **MISSING CORE DOMAIN PATH**.

## BP-J03 — build/resolve artifacts → deploy exact application

**Outcome:** each component resolves to verified prebuilt/built/non-artifact resource; runtime receives exact immutable subject and returns ProviderResource identities.

Current:
- live Go handler sends unrelated mutable `alpine:latest`;
- BuildEngine adapter contract exists only in specification/prototype;
- delegated selected-app build prototype pending.

Classification: **CURRENT LIVE PATH CONTRADICTS MATURE JOURNEY**.

## BP-J04 — uncertain provider operation → restart/reconcile

**Outcome:** provider mutation is journaled before side effect; timeout/crash/lost response becomes UNKNOWN then reconciles without duplicate/leaked resources.

Current:
- BP-F04 natively reproduces remote-success/local-persistence-failure risk;
- no production operation journal;
- executable reference model pending;
- duplicate request probe pending.

Classification: **NATIVE-RISK-REPRODUCED / JOURNEY ABSENT**.

## BP-J05 — observe exact realized deployment → stop/update/rollback

**Outcome:** operator observes the exact generation/resources/application identity and can stop/update/rollback intended resources only.

Current:
- project stores provider sandbox IDs;
- BP-F03 natively proves terminate uses ProjectID instead;
- no mature generation/rollout model implemented;
- selected application observation missing.

Classification: **WRONG IMPLEMENTATION / NOT CLOSED**.

## BP-J06 — deployment facts → portfolio projection → independent publication

**Outcome:** verified source/deployment facts produce a projection separating verified/generated/manual content; publisher updates destination independently of deployment truth.

Current:
- direct user-backed mature intent recovered;
- historical Demonstrator implementation corroborates shape;
- current named portfolio implementation path absent;
- publisher adapter not implemented.

Classification: **MATURE OBLIGATION / CURRENT IMPLEMENTATION MISSING**.

## BP-J07 — installed CLI/Desktop/API → same durable operation

**Outcome:** user can initiate/observe one domain Operation through supported projections without semantic divergence.

Current:
- desktop/web performs real Go API mutation;
- Tauri CLI deploy is plan-only;
- common application operation not implemented;
- backend binds all interfaces by default.

Classification: **ARCHITECTURAL DRIFT / NOT CLOSED**.

## BP-J08 — update/uninstall local control plane without corrupting remote truth

**Outcome:** local update/rollback/uninstall preserves/migrates operation/resource history and never silently destroys remote resources.

Current:
- distribution surfaces/docs exist;
- v1.1 migration/uninstall semantics not implemented/proven.

Classification: **LIFECYCLE GAP**.

## First vertical slice recommendation

Use a **single disposable selected application**:

1. authenticate;
2. select repository/ref;
3. resolve immutable SourceSnapshot;
4. read/validate exact ManifestRevision;
5. resolve one component through BuildEngine;
6. obtain immutable BuildArtifact identity/provenance;
7. persist RuntimeOperation before mutation;
8. deploy through disposable RuntimeAdapter;
9. persist ProviderResource identity;
10. independently query `/__byteport_probe`;
11. compare source/manifest/artifact identity;
12. restart BytePort control process;
13. reconnect to same operation/deployment;
14. stop exact provider resource;
15. prove unrelated resource survives;
16. derive a realized PortfolioProjection;
17. fail publication deliberately and prove deployment truth remains realized.

Required adversaries:
- branch moves after resolution;
- wrong artifact returns HTTP 200;
- provider response lost;
- local persistence failure;
- same OperationID/different fingerprint;
- delayed provider visibility;
- wrong ProviderResourceID;
- expired session token;
- wrong principal/target;
- publication failure.

Until this closes, BytePort is not a narrow functional mature deployment product.
