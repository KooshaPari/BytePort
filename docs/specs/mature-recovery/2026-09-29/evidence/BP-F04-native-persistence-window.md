# Native evidence receipt — BP-F04 remote side-effect / persistence window

Date: 2026-09-30.

## Probe history

Probe v1 closed the SQLite database before handler execution. It failed to reach the provider and was classified **INVALID PROBE / NOT PRODUCT EVIDENCE**.

Probe v2 injects failure only in GORM's Project create callback. The database remains available while the handler performs the provider request.

## Exact execution

- candidate: `aae3fbaa1915cccb9498635c69e117d71a81763b`
- CI run: `36686241335`
- Go-test job: `109792659013`
- probe: `TestRecoveryProbeRemoteDeploySurvivesLocalPersistenceFailure`

The routes package ultimately failed because the separate BP-F03 adversarial test still fails. The BP-F04 probe itself did not report failure.

Its assertions require all of the following:
1. provider deploy call count == 1;
2. local handler response is server error after injected persistence failure;
3. provider stop/compensation call count == 0;
4. no local Project row exists.

Therefore its successful completion inside the package establishes the intended observation.

## Classification

**BP-F04 = NATIVE RISK REPRODUCED for the current handler path.**

The current ordering permits a provider resource to be accepted while local durable Project state fails, with no compensation in the exercised path.

This is not yet a desired-behavior acceptance failure because the replacement recovery contract remains to be selected. Valid architectures include durable operation journal + reconcile/adopt, provider-scoped idempotent retry, or compensation where safe. The mature contract must preserve explicit UNKNOWN/reconciling state rather than pretending the two systems are atomic.
