# BytePort — oracle/autograder design, pass 1

Date2026-09-29; analyzed source `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`. **NOT a completed or deployed grader.** Full accepted criterion denominator, configuration matrix and numeric quality targets remain unresolved. No generated/inherited catalog is admitted to product grading by this dossier.

## Three lifetimes and the control loop

An ephemeral developer worker attempt (agent/human/tools/worktree/lease) operates on a durable development effort (change intent, plan, dependencies, reviews, grader receipts). Neither is BytePort's product state: projects, configurations, source/build identities, deployment intent, provider resources, releases and observations must persist independently. Replacing the developer worker must not delete its effort history; restarting a BytePort operation worker must not lose deployment truth. Completing a work package is not acceptance of a running application.

Accepted contract → bounded context → worker action → independent verifier → multidimensional findings → localized next context → retry/replan/clarify/escalate. The worker can see the rubric but cannot authorize its own rubric changes or substitute a different candidate's evidence.

## Exact evidence and authority

Require product/subject; accepted contract revision and criterion; candidate source SHA and packaged artifact digest; manifest/source/build/target/operation/provider identities as applicable; configuration; OS/runtime/dependency versions; verifier identity/version/digest; run ID; timestamp; raw artifact location/digest and executed-case count. Collect independent observations of runtime state and requested application identity. Bind authorization evidence to the actual actor/resource/target, not merely an authenticated request.

Wrong candidate/contract/target, stale observations, missing artifacts, failed collector, unknown provider outcome, empty or skipped tests are never green. A structurally valid receipt still is not proof that its observation is truthful. Preserve source authority and validation state on bidirectional intent↔design↔implementation↔test↔evidence↔runtime edges; confidence cannot turn an agent inference into user intent.

## Non-negotiable negative controls for the first slice

| Behavior | Positive control | Adversarial/failure controls | Independent observation |
|---|---|---|---|
| Selected application deploy | Requested source/manifest leads to expected running marker/digest | Wrong repo/image with200; mutable tag changed; unsupported build/target; empty provider response; missing repository | Built digest and actual application response, not only control API response |
| Ownership and provider identity | Owner stops actual returned runtime ID | ProductUUID != providerID; foreign owner; forged body owner; wrong account/region; stale deployment | Captured provider request plus actual targeted/untouched runtime state |
| Retry/recovery | One durable intent reconciles to one known deployment under declared adapter semantics | Lost provider response; DB write failure; crash before/after receipt; timeout; duplicate callbacks; replaced worker | Durable journal and provider lookup; uncertain state remains explicit |
| Dependency failure | Clear bounded error with preserved recovery state | Runtime unavailable, bad auth, keyring locked, corrupt storage, partial migration, disk full | Actual error/side effect and restart result, not a log substring |
| Desktop access | Supported window/session reaches intended scoped service | Wrong origin/window, external network access in local-only mode, untrusted content, expired credentials | Installed Tauri/API path and actual permission/bind behavior |
| Presentation | Current permitted deployment facts become reviewable output | Stale/wrong project's facts, failed deploy presented live, secret leakage, hallucinated claim, unauthorized publish | Provenance links and exact rendered/exported content |
| Regression and lifecycle | Supported version upgrades without losing state | Old schema, rollback incompatibility, stale signed artifact, uninstallation ambiguity | Versioned fixtures, signed artifact identity and recovered state |

Mutation questions: does replacing the selected image with alpine still pass? Can equal-ID fixtures conceal the stop bug? Can deleting persistence/authorization/identity checks manufacture green? Can the candidate change the grader to skip desktop/remote-state checks? Can a200 or healthy control server hide an absent app? Every such shortcut needs an independently owned counterexample.

## Independence, progress and quality

Seal accepted grading policy separately from candidate source, require authorized review for policy changes, and retain immutable raw run history. Candidate adapters can expose observations but cannot declare their own eligibility or acceptance. Test real mounted interfaces and actual persistence; mocks supplement but cannot replace public-path acceptance. The enforcement system for this separation is still unimplemented in this pass.

Report functional/trace/evidence/journey/regression/performance/reliability/security/accessibility/usability/uncertainty/transition-debt dimensions separately. Security, wrong-artifact and wrong-target failures veto acceptance rather than averaging away. Scope growth and engineering improvement are separate deltas. Evidence age and environment are visible. Velocity, stagnation and oscillation require comparable repeated observations; none is inferred here.

Latency/recovery/cost/availability targets need accepted workloads and explicit thresholds. They are not copied as generic requirements onto every feature. Baseline product/alternative/status-quo pilot comes after a usable built stage; it cannot substitute for this pre-build design gate.
