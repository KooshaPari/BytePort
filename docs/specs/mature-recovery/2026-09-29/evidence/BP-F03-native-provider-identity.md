# Native evidence receipt — provider runtime identity

Date: 2026-09-29.

## Subject
BytePort termination must target the persisted provider runtime identity rather than the BytePort project identity.

## Candidate
- Product repository: `KooshaPari/BytePort`
- Adversarial test commit: `934b0456de5509b5e89294c3adaa046e632a2767`
- Test: `backend/byteport/routes/recovery_provider_identity_test.go::TestTerminateInstanceUsesPersistedProviderSandboxID`
- Workflow run: `36615383712`
- Re-run job: `109572633061`
- Job: `Go test (backend/byteport)`

## Environment/result
The exact job reached `go test`; unrelated packages passed. The routes package failed on the adversarial test with:

`stop id = "project-123", want persisted provider id "provider-sandbox-777"`

The fixture deliberately separates:
- BytePort project identity: `project-123`
- provider runtime identity: `provider-sandbox-777`

The handler deleted the project after successfully issuing the stop request against the wrong ID, so a 200-like provider stub alone would have permitted a false green without inspecting the actual target.

## Classification
**NATIVE COUNTEREXAMPLE REPRODUCED / BP-F03 CONFIRMED FOR THIS CANDIDATE.**

This establishes the wrong-target identity defect for the exact candidate and handler path. It does **not** establish:
- every provider/runtime implementation is affected;
- selected-source deployment correctness;
- crash-window recovery correctness;
- remediation;
- an accepted final ontology;
- overall product failure or completion percentage.

## Architecture consequence
ProjectID and ProviderResourceID are distinct first-class identities. Termination/reconciliation criteria must bind to the realized deployment/provider resource, with ownership and generation/operation provenance. A fix that merely happens to select a map value is insufficient until multi-deployment selection semantics are defined.
