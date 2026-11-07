# BytePort — pass 6 lineage correction: portfolio generator and NanoVMS ejection

Date: 2026-09-30. Frozen analyzed product source remains `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`.

## Historical template/generator implementation did exist

Git history contains commit `c1ceea9773b57005419d166a696ef84e1875a043` (2024-12-24 UTC), titled:

> Extended MVP With Template Engine - Proj Playground Next / Local LLama

The commit contains extensive historical work under:
- `.history/backend/nvms/Demonstrator/`;
- historical BytePort integration/auth paths;
- OpenAI-related demonstrator generation experiments.

The surviving historical generation files are fragmentary and implementation-shaped. They prove experimentation/implementation lineage, not that the current mature contract should restore those exact modules.

Classification: **HISTORICAL IMPLEMENTATION / EXPERIMENT supporting direct user productization intent**.

Recovery principle:
- recover the desired portfolio projection/publishing outcome and useful schema/interaction semantics;
- do not resurrect the historical Demonstrator/NVMS implementation wholesale merely because code once existed.

## NanoVMS removal was repository-boundary consolidation, not product-scope deletion

Commit `e81c90f6600f76325e2066bf4e1ac6754c3ab93b` says explicitly:

> NVMS was ejected to nanovms repo (merge of consolidate/B11-delete-nvms-impl).

It removes stale `backend/nvms` references from CI. Follow-up `6457c2e208c0b72efe26c29119b6aae9fa6499e6` is an empty CI trigger to verify “nvms-free CI.”

This is strong historical evidence that:
- BytePort stopped owning the NanoVMS implementation;
- the runtime/sandbox capability moved behind a repository boundary;
- removing `backend/nvms` from BytePort does **not** by itself remove BytePort's deployment/productization obligation.

The standalone/canonical NanoVMS remote is not currently readable through the expected public GitHub path in this environment, so its current source/API cannot be assumed. Integration must be verified against an actual accessible provider contract before architecture freeze.

## 2026-09 duplicate-backend retirement is a separate cleanup

Commit `aba00baf2eb57faf8b7ec6ee487ba0ed085a2b0c` retires:
- the divergent `github.com/byteport/api` backend;
- the `bytebridge` Fermyon experiment;
- associated duplicate models/internal/cloud code.

Its commit message explicitly says `backend/byteport` remains canonical and that no repository consumer reached the removed duplicates.

Classification: **IMPLEMENTATION CONSOLIDATION**, not mature-scope authority.

## Portfolio Git fallback implementation status

Current-tree searches for Git publication found:
- GitHub OAuth/token validation and repository listing;
- repository validation;
- developer/release/CI git-push utilities.

No current product path was found that publishes portfolio projections to a target repository.

Therefore the recovered user-authorized Git/server-side portfolio fallback is:
- **MATURE OBLIGATION CANDIDATE backed by direct user intent**;
- **NOT CURRENTLY IMPLEMENTED**;
- and must use a separate publication-target authorization model rather than assuming the source repository is the publication repository.

## Lineage consequence

The mature architecture should preserve clean boundaries:

`BytePort domain/control plane`
→ delegated build engine
→ runtime/provider adapter (NanoVMS is one historical/current candidate)
→ observed realized deployment
→ BytePort PortfolioProjection
→ external PortfolioPublisher adapter / authorized Git fallback

Historical Demonstrator/NVMS code informs behavior and failure cases but does not dictate module ownership.

## Remaining falsification

- recover any surviving portfolio payload/template schema from historical Demonstrator/Slickport artifacts;
- inspect accessible NanoVMS API/source if/when a canonical source can be resolved;
- establish whether any user decision superseded the Git publication fallback;
- test generic PortfolioPublisher behavior independent of the retired implementation.
