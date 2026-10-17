# BytePort — pass 5 conversation authority: deployment-to-portfolio integration

Date: 2026-09-30.
Status: authority recovery; implementation and exact external portfolio owner remain to be mapped.

## Direct user-authored portfolio intent

### 2024-11-24 — deploy plus portfolio projectization
The user described BytePort deploying a project as its own Docker instance on AWS, then using AI to generate a portfolio project page with live access/descriptions and adding a list-style projects entry.

This establishes that portfolio/productization is part of BytePort's original mature outcome, not a later documentation embellishment.

### 2024-11-24 — template/API integration model
The user specified a template-oriented integration:
- fetch pre-existing project-list and project-page components/templates via an API;
- fill/push completed dynamic JSON objects;
- use AI-generated project data plus separately generated/mockup modifications;
- retain programmatic Git/server-side processing as fallback;
- make template/API location discoverable or directly handed off.

The user accepted the hybrid direction and deferred implementation.

### 2024-11-25 — integration is non-negotiable
The user explicitly required the flow to:
- parse NVMS;
- provision AWS VM/environment resources;
- GET portfolio templates for list and full-project pages;
- generate portfolio content with OpenAI;
- POST the result back;
- fall back to programmatic Git upload if the API is unavailable;
- provide project/instance monitoring.

### 2024-11-24/25 — configured external portfolio endpoint
User/account-flow context included Portfolio.RootEndpoint/APIKey together with Git, AWS and OpenAI credentials, and implementation discussion persisted the portfolio URL/API key.

**Authority consequence:** the original model is not simply “BytePort renders its own portfolio page.” BytePort owns deployment→portfolio projection/generation/integration, while a configured external portfolio endpoint/site owns at least part of rendering/storage. Git-based publication is an explicit fallback path.

## Portfolio ontology correction

Replace the earlier unresolved question “render vs publish vs emit” with the stronger provisional model:

### PortfolioProjection
BytePort-owned, evidence-derived project representation tied to:
- Project;
- SourceSnapshot;
- ManifestRevision;
- RealizedDeployment / endpoint observations;
- generated descriptive/media metadata with provenance.

### PortfolioPublisher
Adapter to a configured external portfolio surface:
1. template/API fetch;
2. JSON/project payload publication;
3. Git/server-side fallback when API path is unavailable.

The external renderer may historically be Slickport/Slick or another configured portfolio implementation. The exact historical system still needs repository/lineage proof; the BytePort obligation is the integration contract, not ownership of an entire portfolio CMS.

## Mature horizon impact

Portfolio/productization is promoted from “candidate mature behavior” to **direct-user-intent-backed mature behavior**, subject to exact acceptance decomposition.

The current repository's missing named J3 implementation paths therefore represent a real implementation gap, not evidence that portfolio scope was intentionally removed.

## Surface authority remains unresolved

No recovered direct user statement in this pass authorizes a mature pivot from cloud/deployment/productization to local-only desktop, nor a definitive CLI-first→desktop-first replacement.

Thus:
- local-first may describe control-plane placement/CVP;
- desktop is a valid implemented surface;
- CLI is a valid accepted/historical surface;
- mature product identity remains deployment + productization;
- CLI-vs-desktop priority remains a projection/UX decision, not a product-identity fork.

## Next lineage work

- identify the exact historical Slickport/Slick repository/API contract if recoverable;
- map current portfolio credential/settings surfaces to the original integration contract;
- decide whether the Git fallback remains mature-mandatory or becomes compatibility/contingency behavior;
- derive portfolio provenance/privacy/secret-redaction acceptance cases.
