# BytePort implementation mapping — mounted surfaces pass 2

Date: 2026-09-30. Frozen source `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`.

## Mounted Go API

`setupRouter()` mounts:
- authenticated link/authenticate/instances/projects/repository discovery/deploy/terminate/credentials;
- public login/signup/GitHub callback/health/metrics.

The server starts on:

`0.0.0.0:<resolved port>`

not loopback.

CORS permits selected localhost/0.0.0.0/Tauri/emulator origins. CORS is a browser policy and does not make the API local-only or authenticated by network placement.

## Network-boundary contradiction

The recovered local-first/desktop-control-plane framing cannot be represented as “local-only” while the backend unconditionally binds all interfaces.

This does not prove the current service is remotely exploitable; protected routes use AuthMiddleware. It does prove network reachability and application authentication are separate concerns and the deployment mode is not currently explicit.

Required mature configuration:
- loopback/local mode;
- explicitly authorized LAN/tailnet mode where supported;
- externally reachable mode only if intentionally supported with corresponding auth/TLS/proxy policy.

Default mode requires an accepted decision.

## Tauri/headless CLI

Actual Tauri binary mounts:
- desktop;
- plan-only `deploy --target`;
- transport check/status;
- version.

No mounted CLI execution path currently invokes the Go deployment domain operation. Therefore accepted CLI-first authority is not currently realized as CLI deployment capability.

## Surface-to-ontology mapping

| Surface | Relation | Current mature disposition |
|---|---|---|
| web project/repository UI | Project/SourceReference projection | partial |
| web `POST /deploy` | provider mutation | live but semantically incomplete |
| web `POST /terminate` | provider lifecycle | live; BP-F03 native defect |
| repository discovery | SourceReference | useful precursor; no immutable snapshot |
| Tauri `deploy` | plan presentation | scaffold, not deployment |
| Tauri transport | old transport lineage/integration | must justify against mature deployment/productization scope |
| desktop shell | human projection | first-class peer surface |
| public health/metrics | operational observation | needs exposure/security policy |
| portfolio | no mounted mature current path found | mature obligation/current gap |

## New falsification questions

1. Which default network mode is accepted for desktop/local-first operation?
2. Can the CLI create/observe the same domain Operation as desktop/API without a second deployment implementation?
3. Are public health/metrics intended to remain unauthenticated in every network mode?
4. Does transport upload functionality serve BuildEngine/artifact publication, or is it residue from the superseded transport lineage?
5. Can installed desktop start/locate/authenticate the backend without relying on globally reachable `0.0.0.0`?

## Consequence

BytePort currently has a **real networked Go control plane plus a desktop shell and plan-only Rust CLI**, not a purely local desktop application. Mature architecture must make placement/network mode explicit and converge all projections on one application operation.
