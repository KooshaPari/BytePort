# Architecture

## Overview
- BytePort is a multi-language workspace centered on the Tauri frontend and backend services.
- The root Go modules and frontend assets support the desktop experience and supporting automation.
- This document is a skeleton for filling in the authoritative architecture details.

## Components
## frontend/web/src-tauri
- Desktop application shell and Rust-side integration for the frontend.

## backend/byteport
- Go backend service for BytePort domain logic and server-side coordination.

## backend/nvms
- Supporting Go module for BytePort infrastructure or platform integration.

## frontend/web
- Web frontend assets and application code for the user-facing experience.

## Data flow
```text
user actions -> frontend/web -> frontend/web/src-tauri -> backend services -> external systems
```

## Key invariants
- Keep the frontend shell and backend services aligned on shared contracts.
- Prefer explicit interfaces between the Tauri layer and Go services.
- Do not bypass the documented service boundaries when adding new features.

## Cross-cutting concerns (config, telemetry, errors)
- Config: keep runtime settings centralized and environment-driven.
- Telemetry: standardize logs and traces across the desktop and backend layers.
- Errors: surface actionable failures at the boundary layer and preserve context internally.

## Future considerations
- Replace placeholders with crate-level responsibilities and ownership.
- Document persistence, sync, and packaging boundaries once they stabilize.
- Add diagram detail for startup, sync, and release flows.
