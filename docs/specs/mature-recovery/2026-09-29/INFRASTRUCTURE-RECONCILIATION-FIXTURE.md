# BytePort generalized infrastructure reconciliation fixture — v0.1

Date: 2026-09-30.
Status: **required before BP-WP-B08 production integration**.

## Goal

Prove the desired-resource graph and reconciliation semantics on heterogeneous resource kinds without coupling to a real cloud or bare-metal provider.

## Desired graph fixture

At minimum:
- runnable service resource;
- provider-managed database-like resource with no BuildArtifact;
- network-like resource;
- dependency edge service -> database/network;
- explicit Target;
- stable resource config digests.

## Observed graph cases

1. resource absent -> CREATE;
2. exact fresh match -> NOOP;
3. fresh mutable mismatch -> UPDATE;
4. mismatch requiring immutable replacement -> REPLACE;
5. stale/incomplete observation -> UNKNOWN;
6. realized resource absent from desired graph, no destruction intent -> READ/ORPHAN;
7. same with exact authorized DestructionIntent -> DELETE;
8. unrelated realized resource -> never touched;
9. provider capability missing -> explicit unsupported/defer, not approximation.

## Restart/reconciliation

Persist:
- DesiredResourceGraph revision;
- ResourcePlan;
- Runtime/Infrastructure Operation identity;
- RealizedResource references.

Recreate controller and prove the next reconciliation uses the same identities rather than creating parallel state.

## Target capability fixtures

Compare at least:
- cloud-like target capabilities;
- bare-metal-like target capabilities.

Do not require a real provider yet. The fixtures must differ meaningfully, e.g. bare metal supports inspect/provision/clean but not live migration.

## Destruction safety

DELETE requires:
- exact DesiredResource/RealizedResource identity;
- current observation;
- explicit authorized DestructionIntent;
- adapter capability;
- reason in plan.

Local BytePort uninstall/deletion of local metadata is never sufficient.

## Evidence

Receipt binds desired graph digest, observed graph, target capabilities, policy, planner version, actions/reasons, candidate SHA and restart result.

## Promotion

B08 may proceed to real provider integration only after this fixture passes and SOTA/bootstrap review remains consistent with the planner abstractions.
