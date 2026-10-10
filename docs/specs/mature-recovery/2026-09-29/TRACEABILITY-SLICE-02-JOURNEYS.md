# Traceability slice 02 — selected-app vertical journey

Date: 2026-09-30. Supplements TRACEABILITY-SLICE-01.

| Trace ID | Journey criterion | Current source relation | Oracle/contract | State |
|---|---|---|---|---|
| BP-T13 | requested ref resolves immutable SourceSnapshot | Repository model has mutable metadata/default branch only; deploy request carries repository/id | BP-VS-01 | missing |
| BP-T14 | exact ManifestRevision before mutation | no mounted manifest path in deploy handler | BP-VS-02 | missing |
| BP-T15 | component ArtifactResolution | no mounted component/build domain | BP-VS-03 | missing |
| BP-T16 | immutable BuildArtifact/provenance | current handler uses mutable unrelated alpine:latest | BP-VS-03/04 | conflicting implementation |
| BP-T17 | RuntimeOperation persists before mutation | current provider call precedes Project save; BP-F04 | BP-VS-05/08/09/10/11 | native risk reproduced |
| BP-T18 | ProviderResource typed identity | sandboxResp.ID stored inside Project deployments map | BP-VS-06/12 | partial; stop path wrong natively |
| BP-T19 | selected application independently observed | provider response/access URL not source/artifact probe | BP-VS-07 | missing |
| BP-T20 | rollout generation | no typed generation model | BP-VS-13 | missing |
| BP-T21 | auth session lifetime enforced | static ValidateToken concern; native expired-token oracle queued | BP-VS-14/15 | pending evidence |
| BP-T22 | portfolio projection/publication | mature direct intent + historical evidence, current path absent | BP-VS-16/17 | missing |
| BP-T23 | explicit network mode | server binds 0.0.0.0; CORS + AuthMiddleware separate | BP-VS-18 | configuration/architecture gap |
| BP-T24 | CLI/Desktop/API common operation | web mutates; Tauri deploy plan-only | BP-VS-19 | architectural drift |
| BP-T25 | uninstall remote-resource disposition | lifecycle obligation exists; public implementation unproven | BP-VS-20 | open |

Storage relation:
Current Project + DeploymentsJSON cannot represent BP-T13..T20 without additive identities. Legacy rows therefore remain `legacy_unverified` unless exact historical evidence qualifies them.

Reverse path example:
`projectData.deployProject → POST /deploy → Project repository metadata → alpine:latest → sandboxResp.ID map → BP-T13..T19 → VERTICAL-SLICE-SELECTED-APP`.
