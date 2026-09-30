# BytePort portfolio/Slickport lineage — pass 2

Date: 2026-09-30. Related-repository archaeology only; Slickport is **not** started as a third product program.

## Authority recovered

### Direct user intent — November 2024

User-authored BytePort context establishes portfolio/productization as part of the original mature outcome:

- deploy a project as its own AWS-hosted runtime;
- generate a portfolio project page with live access/descriptions;
- add/update the list-style projects entry;
- fetch pre-existing project-list and full-project templates/components through an API;
- generate project data/content with AI and fill dynamic JSON/template objects;
- POST the resulting project representation back to the configured portfolio surface;
- retain programmatic Git/server-side publication as fallback when the API path is unavailable;
- carry a configured portfolio root endpoint and API key alongside Git/AWS/LLM integration;
- expose project/instance monitoring so published/live project state has an operational source.

This is **USER INTENT**, not an assistant-generated feature expansion.

The exact rendering implementation was not required to live inside BytePort. The recovered model is integration-oriented: BytePort transforms verified project/deployment state into a portfolio representation and publishes it to a configured external portfolio system.

## Slickport evidence

The BytePort charter treats portfolio integration as first-class and describes the portfolio backend as “e.g. Slickport.” The historical embedded README says demo portfolio integration expected Slickport credentials.

PhenoRegistry historical inventories list `KooshaPari/slickport` as a GitHub-only product/app and separately as “Keep (pending triage).” Current direct repository lookup returns 404 and current repository search does not surface it.

Classification:
- Slickport = **historical portfolio backend / concrete integration example**;
- Slickport is **not** BytePort product identity;
- present GitHub source = unavailable through current repository lookup;
- portfolio integration remains normative because its authority is independently recovered from user-authored BytePort intent.

## Mature product boundary

### PortfolioProjection — BytePort-owned

A projection is derived from:
- Project;
- immutable SourceSnapshot;
- ManifestRevision;
- RealizedDeployment;
- fresh endpoint/runtime Observation;
- generated descriptive/media metadata with explicit provenance.

Verified facts and generated copy are separate fields/authority classes.

### PortfolioPublisher — adapter boundary

The publisher supports the recovered integration pattern:

1. discover/fetch target portfolio templates/component contracts;
2. produce the project-list and project-page dynamic representations;
3. submit structured project payloads to the configured portfolio API;
4. preserve the API/root-endpoint identity in publication evidence;
5. fall back to an authorized Git/server-side publication path when the API path is unavailable;
6. observe publication result separately from deployment success.

A future static-site, CMS, Slickport-like backend, or other portfolio implementation can satisfy this adapter contract. BytePort does not need to absorb a portfolio CMS.

## Required acceptance/negative controls

- wrong Project/RealizedDeployment cannot populate another project's portfolio entry;
- stale/dead endpoint cannot be represented as freshly verified live;
- generated prose cannot be labeled as verified deployment fact;
- secrets, credentials and private environment values never enter projection/publisher payloads;
- API publication failure cannot turn deployment status red if deployment itself succeeded;
- Git fallback must target the authorized repository/path and preserve provenance;
- API recovery/retry must not duplicate project entries;
- publication of a new deployment must not silently erase historical deployment identity;
- operator can distinguish generated, verified, stale and manually reviewed fields;
- unavailable historical Slickport must not block portfolio export through another accepted adapter.

## Implementation-state implication

Current named J3 implementation paths are absent from the frozen implementation. That is a **real product gap**, not evidence of intentional scope deletion.

Current portfolio credential/settings remnants are lineage evidence, but existence of fields is not journey closure.

## Remaining archaeology

- recover any surviving Slickport API/template schema from registry snapshots, historical BytePort files or conversations;
- determine whether the “all three methods” historical hybrid meant API template retrieval + structured publication + Git fallback in every deployment or as layered contingencies;
- map current user/settings credential fields to the mature publisher adapter;
- derive privacy/review/publication quality overlays;
- test a generic static/content adapter before any custom portfolio backend work.

No Slickport code is imported and no third-repository branch is created.
