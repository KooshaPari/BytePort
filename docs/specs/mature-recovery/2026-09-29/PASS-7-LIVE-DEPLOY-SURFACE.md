# Pass 7 — live deploy surface implementation mapping

Date: 2026-09-30. Frozen implementation source remains `0232cca16fedb7963a8c6f556dc5eee5c8c1674e`.

## Web UI → mounted backend

`frontend/web/src/components/projectData.ts::deployProject` is the live web caller for `POST /deploy`.

It sends:
- name;
- description;
- type;
- platform;
- repository object;
- empty readme.

It does **not** send:
- requested branch/tag/ref;
- resolved commit SHA;
- manifest bytes/path/digest;
- build-plan identity;
- artifact digest;
- product operation/idempotency ID;
- target/provider identity beyond loose platform/type metadata.

The mounted Go handler accepts this shape, persists repository identity, then constructs a NanoVMS SandboxConfig containing hard-coded `alpine:latest`.

Therefore the current live UI journey cannot establish FR-DEPLOY-003/004 or the recovered mature source→manifest→artifact contract.

## Tauri/headless deploy surface

`frontend/web/src-tauri/src/deploy.rs` defines `byteport deploy --target` as a **side-effect-free planner**.

Its own source states:
- it materializes a compiled-in stage catalog;
- renders JSON;
- executes no stage.

The stage commands are future-looking strings such as:
- `byteport target verify`;
- `byteport artifact build`;
- `byteport config apply --plan-only`.

Thus the Tauri/headless `deploy` surface is not an alternate implementation of the mature deployment journey. It is a planning/proposal surface.

## GitHub repository discovery

`GET /api/github/repositories` retrieves repositories using the authenticated user's stored Git token. This establishes repository discovery/auth plumbing but not source snapshot resolution. No inspected mounted deployment path takes the selected repository and resolves a requested branch/ref to an immutable commit before provider mutation.

## Implementation truth table

| Mature stage | Web UI / mounted path | Tauri CLI | State |
|---|---|---|---|
| Project metadata | yes | plan target only | partial |
| Repository discovery | yes | no mature source selection | partial |
| SourceReference | repository object/id only | no | incomplete |
| SourceSnapshot | absent | absent | missing |
| ManifestRevision | absent | absent | missing |
| BuildPlan | absent in mounted deploy; placeholder hard-code | static plan strings only | proposal/scaffold |
| BuildArtifact | unrelated mutable `alpine:latest` | not executed | incorrect/missing |
| Operation identity | absent | not executed | missing |
| ProviderResource | returned sandbox ID stored | not executed | partial |
| Application observation | not tied to selected source/artifact | not executed | missing |
| Portfolio projection | mature intent; current named journey implementation absent | absent | missing |

## Consequence

BytePort currently has two disconnected partial surfaces:
1. a real web/backend path that performs provider mutation without selected-source/build identity;
2. a Rust planner that describes future deploy stages without executing them.

Do not claim either as a closed deployment vertical slice.

The proposed delegated BuildEngine architecture should unify these by making both UI and CLI project onto the same application/domain operation, rather than maintaining a planning CLI and unrelated mutation API.

## Required vertical-slice mapping

A future CVP slice must trace one exact public action through:

`UI/CLI → SourceReference → SourceSnapshot → ManifestRevision → BuildOperation → immutable BuildArtifact → RuntimeOperation → ProviderResource → live application observation → durable evidence`.

Every transition must be mounted/reachable and independently evidenced.