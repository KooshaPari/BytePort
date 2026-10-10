# BytePort unresolved-decision register — non-code finality

Date: 2026-10-01.

| ID | Decision | Why unresolved | Required closer | Default until closed |
|---|---|---|---|---|
| BP-UD-01 | default control-plane network mode | product/operational preference not directly settled | direct authority + threat/reachability experiment | loopback/local safest default for prototypes; no mature freeze |
| BP-UD-02 | first bare-metal engine | Ironic/MAAS/Tinkerbell operational fit empirical | disposable comparison | adapter-neutral; Ironic+Bifrost leading experiment |
| BP-UD-03 | BuildEngine first-party support set | portability/maintenance empirical | build/provenance comparison | interface only; no brand identity |
| BP-UD-04 | provenance mechanism | attestation implementation not verified | BuildKit/OCI provenance experiment | artifact identity without attestation is not provenance-complete |
| BP-UD-05 | richer manifest v1 fields | historical schema authority incomplete | Claude/history/user authority + implementation experiment | recovery-v1 minimal schema only |
| BP-UD-06 | portfolio publication destinations/policy | product preference and credentials/site policy | explicit publication experiment/authority | projection independent; publication optional |
| BP-UD-07 | numeric SLOs | historical ADR numbers unqualified | benchmark + accepted target | metrics informational |
| BP-UD-08 | provisioning vs host-configuration default adapter | site/OS dependent | Nix/Ansible/cloud-init comparison | separate capability boundary; no core config DSL |

Everything else is governed by the v1.2 generalized infrastructure contract.
