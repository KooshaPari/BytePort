# Architecture decision candidate BP-AD-03 — CLI-primary domain, desktop peer projection

Status: **PROVISIONAL ACCEPT / direct-user supersession check remains open**  
Date: 2026-09-30.

## Authority basis

The frozen accepted ADR set contains ADR-005 — **CLI-First Interface**.

Later September work adds a real Tauri/Svelte desktop shell, but direct commit archaeology classifies the README desktop-market-positioning rewrite as agent-authored documentation and the application-shell commit as an implementation/UI evolution. No direct user statement or executive product-authority decision has been recovered that replaces the CLI-first decision.

Therefore document recency is insufficient to invert interface authority.

## Decision

BytePort has **one application/domain contract** and multiple projections.

- CLI remains the primary normative automation/operator projection unless later direct authority supersedes ADR-005.
- Desktop is a first-class human projection over the same domain operations.
- Web/Tauri presentation must not own separate deployment semantics.
- Machine/API interfaces expose the same operation identities and evidence model.
- Interface differences may affect interaction design, not Project/SourceSnapshot/ManifestRevision/BuildArtifact/Operation/ProviderResource identity.

## Current-state implication

The present surfaces are not yet projections of one operation model:

- web UI calls mounted `POST /deploy`, which performs provider mutation using placeholder artifact identity;
- Tauri/headless `byteport deploy --target` is explicitly plan-only and executes no stages.

This is architectural drift, not evidence for two intended products.

## Required convergence

A mature deployment action should exist once in the application/domain layer.

CLI:
`byteport deploy ... → create/observe Operation`

Desktop:
`Deploy button → create/observe the same Operation`

Machine API:
`POST/command → create/observe the same Operation`

Each projection may:
- select source/target;
- review a plan;
- authorize execution;
- stream/display status;
- request stop/rollback/retry;

but none may invent its own deployment identity or provider semantics.

## Negative controls

- CLI and desktop request the same immutable DeploymentIntent and receive different operation semantics.
- desktop can deploy a source/ref the CLI cannot represent.
- CLI reports plan success while no operation exists and UI presents that as deployment success.
- desktop uses ProjectID as provider ID while CLI uses ProviderResourceID.
- one projection bypasses authorization/evidence requirements.
- retry from one projection duplicates an operation started in another.

## Falsification

Revise this decision if direct user authority is recovered that explicitly makes desktop the sole/primary product surface and intentionally demotes CLI to compatibility tooling.

Even then, the single-domain-operation requirement remains unless product intent explicitly defines distinct deployment products.

## Gate consequence

The CLI-first vs desktop-first ambiguity no longer blocks product ontology. It remains an interface-priority question with a provisional authority answer.

No production routing changes are authorized by this decision alone.
