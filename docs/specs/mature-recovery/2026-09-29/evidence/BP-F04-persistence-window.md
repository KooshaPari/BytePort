# BP-F04 probe status — corrected injection

Date: 2026-09-30.

The first BP-F04 probe was invalid: closing the database prevented the handler from reaching provider deployment, so it did not create the intended remote-success/local-persistence-failure window.

The corrected probe installs a GORM create callback that fails only Project creation. This leaves request parsing and provider deployment reachable, then injects failure at local persistence.

In the full Go package run on commit `aae3fbaa1915cccb9498635c69e117d71a81763b`, the corrected probe did not appear among failures while the known BP-F03 adversarial test did. That is suggestive but **not used as the final BP-F04 receipt** because package failure obscures a clean criterion result.

A dedicated workflow now runs only:
`go test ./routes -run '^TestRecoveryProbeRemoteDeploySurvivesLocalPersistenceFailure$' -count=1 -v`

Until that exact isolated job completes, BP-F04 remains `PROBE_EXECUTED_IN_RED_PACKAGE / ISOLATED_RECEIPT_PENDING`.
