# BytePort selected-app autograder contract — v0.1

Status: **CONTROL PLANE / independent grader specification**  
Date: 2026-09-30.

## Dimensions

- source identity;
- manifest validity;
- artifact/provenance identity;
- operation durability/idempotency;
- provider-resource identity;
- reconciliation/recovery;
- selected-application observation;
- authorization/network boundary;
- rollout lifecycle;
- portfolio truth/publication;
- interface convergence;
- evidence/trace;
- regression;
- uncertainty.

## Hard fail conditions

Any of:
- branch/ref used as if immutable SourceSnapshot;
- mutable tag accepted as final artifact identity where digest available;
- provider HTTP success treated as selected-app proof;
- ProjectID used as ProviderResourceID;
- remote mutation before any recoverable operation identity;
- timeout/error collapsed to definite absence/success;
- same OperationID with different fingerprint accepted;
- legacy deployment silently marked verified;
- expired/unauthorized principal allowed to mutate;
- CORS treated as API authorization;
- plan-only result shown as deployed;
- portfolio generated assertion shown as verified fact;
- uninstall silently destroys remote resource.

## Package grader

For each BP-WP:
1. resolve exact ontology/obligation/decision IDs;
2. validate additive migration/compatibility where relevant;
3. execute positive controls;
4. execute negative/adversarial controls;
5. verify exact subject chain;
6. restart control process where lifecycle applies;
7. inspect raw provider/build/probe receipts;
8. reverse-trace implementation;
9. compare prior accepted candidate;
10. emit multidimensional result.

## Selected-app critical set

BP-VS-01..20 are tracked independently.

No overall vertical PASS while any of:
- VS01 source snapshot;
- VS02 manifest;
- VS03 artifact resolution;
- VS05 durable runtime operation;
- VS06 provider resource;
- VS07 selected app;
- VS08/09 restart/lost response;
- VS10 operation identity;
- VS12 exact stop;
- VS14 authorization

is FAIL or UNKNOWN.

Portfolio/publication can be stage-projected later only if the stage definition explicitly permits it; mature contract still retains VS16/17.

## Anti-gaming

- provider fixture returns different ProjectID/ProviderResourceID;
- wrong image responds HTTP 200;
- branch mutates after resolution;
- delayed provider visibility;
- response loss after create;
- duplicate retry;
- legacy DB migration fixture;
- expired session;
- wrong principal/target;
- publication destination fails;
- unrelated runtime must survive stop;
- grader revision/evidence identity immutable.
