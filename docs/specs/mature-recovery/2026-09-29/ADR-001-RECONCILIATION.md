# BytePort accepted-ADR reconciliation — architecture authority vs implementation fiction

Date: 2026-09-30.  
Source: `docs/adr/ADR-001-architecture.md`, status Accepted, last updated 2026-04-03.

## Authority retained

ADR-001 independently supports mature obligations already recovered elsewhere:
- CLI primary interface;
- HTTP/web secondary management interface;
- Git repository + specified branch/ref deployment;
- manifest parsing/validation;
- multi-service orchestration;
- portfolio generation/publication after deployment;
- modular service/infrastructure separation;
- SQLite development persistence with an intended migration strategy.

These claims are accepted design evidence, not direct user intent by themselves. They gain confidence where direct user conversations/older product lineage corroborate them.

## Implementation fiction / stale examples

ADR-001 names a desired package tree such as:
- `backend/byteport/cmd/deploy.go`;
- `internal/manifest/parser.go`;
- `internal/deploy/orchestrator.go`;
- `internal/portfolio/generator.go`;
- `internal/aws/*`.

The frozen current implementation does not match that architecture. The live route remains the Gin handler path already mapped, while Tauri deploy is plan-only.

Therefore:
- ADR package paths are **accepted/proposed architecture**, not current implementation evidence;
- tests cannot cite those paths as proof unless they exist/reach the public path;
- the mature recovery may preserve the semantic service boundaries without recreating the exact package tree.

## Numeric targets quarantined

ADR-001 contains targets including:
- CLI startup <100 ms;
- manifest parsing <500 ms;
- deployment <135 s;
- backend memory <200 MB;
- 5+ concurrent deployments.

No source in the current recovery has yet established:
- measurement methodology;
- workload size/distribution;
- hardware/environment;
- alternative baseline;
- user-operability threshold;
- whether these remain current accepted objectives.

Classification: **historical accepted-design targets / not current mature quality gates until requalified**.

They may become hypotheses/baselines in the quality program, but must not be copied into every requirement or used to grade current candidates without a benchmark contract.

## Threat-model reconciliation

The June threat model describes a production architecture including `backend/nvms`, AWS orchestration and LLM portfolio generation. Several of those paths are absent/stale in the frozen current tree.

Useful retained obligations:
- expiry claim should be mandatory for session auth;
- manifest input is a security boundary;
- cloud/provider credentials need least privilege;
- portfolio/LLM context must exclude secrets;
- Tauri capability scope matters;
- public endpoints/network reachability need explicit security policy.

Implementation assertions in the threat model require source verification before acceptance.

## Consequence

BytePort recovery should preserve the **semantic architecture** that survived independent authority/evidence:
`CLI/API projection → manifest/source services → build/runtime adapters → portfolio service`

without cargo-culting an obsolete directory tree or historical performance numbers.
