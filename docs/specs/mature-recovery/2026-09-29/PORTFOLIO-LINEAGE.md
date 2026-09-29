# BytePort portfolio/Slickport lineage — pass 1

Date: 2026-09-29. This is related-repository archaeology only; Slickport is **not** started as a third product program.

## Evidence

The BytePort charter treats portfolio integration as a first-class mature goal and describes the target persona as running a portfolio backend “e.g. Slickport.” The historical embedded README says demo portfolio integration expects Slickport credentials.

PhenoRegistry inventories `KooshaPari/slickport` as a public/GitHub-only product/app and separately marks it “Keep (pending triage).” Direct GitHub repository lookup now returns 404 and repository search does not surface it.

Classification:
- Slickport = **historical related product / integration example**, not automatically a current dependency;
- current availability = **not accessible through present GitHub repository lookup**;
- BytePort portfolio obligation survives independently because charter/PRD/FR/user-history evidence does not depend solely on Slickport's existence.

## Product-boundary implication

The mature BytePort contract should define a **PortfolioProjection** output/interface rather than make a particular portfolio repository its identity.

Candidate contract:
1. derive projection from authorized Project + SourceSnapshot + RealizedDeployment + fresh Observation;
2. include provenance/freshness and distinguish generated copy from verified facts;
3. redact secrets and private deployment data;
4. support review before public publication;
5. publish/export through an adapter;
6. treat Slickport, a static site, or another portfolio backend as adapters.

This preserves the recovered user outcome—no manual portfolio upkeep—without coupling BytePort's core model to a currently unavailable historical repository.

## Remaining archaeology

- recover deleted/private Slickport history if accessible through registry/deleted-repo records;
- search prior conversations for whether BytePort was meant to absorb Slickport or only integrate;
- inspect any current portfolio API settings/credentials surfaces;
- compare generic static-content/export approaches before building a custom portfolio backend.

No Slickport code is imported and no third-repository specification branch is created.
