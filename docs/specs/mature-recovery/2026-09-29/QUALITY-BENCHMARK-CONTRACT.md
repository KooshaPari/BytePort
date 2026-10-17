# BytePort quality and benchmark contract — v1.0

Date: 2026-10-01.

## Benchmark subject identity

Bind:
- repository/source snapshot;
- manifest revision/desired graph;
- build engine/version;
- artifact/provenance;
- target adapter/provider/version;
- target environment;
- operation/generation;
- BytePort candidate/dependencies;
- observations/raw provider receipts.

## Required baselines

Where meaningful:
- status quo/manual or existing deployment path;
- best realistic alternative stack;
- BytePort candidate.

Bare-metal comparison additionally measures Ironic+Bifrost, Tinkerbell and/or MAAS on an equivalent lifecycle subset.

## Correctness/hard gates

- exact source/manifest/artifact identity;
- no wrong provider resource;
- no duplicate mutation on retry;
- restart/lost-response reconciliation;
- desired/realized graph consistency;
- unauthorized/destructive action rejection;
- UNKNOWN preserved when evidence insufficient;
- local uninstall does not destroy remote resources;
- secrets absent from evidence/publication.

## Operational metrics

- setup/bootstrap effort;
- time to first managed target;
- plan latency;
- build time;
- apply/reconcile time;
- update/rollback time;
- destroy/cleanup time;
- provider API calls/retries;
- human interventions;
- state drift/recovery;
- control-plane resource footprint;
- adapter-specific dependencies/services;
- failure recovery burden.

## Target-setting rule

No historical ADR number is normative without requalification. Numeric SLOs require direct user objective, measured baseline + accepted target, interoperability constraint or safety/operability threshold.

## Heterogeneous-target comparison

Do not compare unsupported semantics as failures. Define portable lifecycle subset, then separately record provider-specific capability breadth.

## Acceptance

BytePort may select different adapters by target/site. No single bare-metal engine must be globally preferred if adapter composition provides the mature contract.
