# BytePort Overhaul Plan

## Session: `byteport:oct-2025-merger`

**Date:** 2026-04-02  
**Status:** Planning

---

## OPTIMAL CHOICES (Most Robust)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **nvms Integration** | **Dual: Library + CLI/IPC** | Library for same-process control, CLI over IPC for cross-platform/cross-language |
| **IaC Tool** | **Pulumi** | Full language IaC, cross-cloud, better testing, mature SDKs |
| **Database** | **PostgreSQL (production) + SQLite (local)** | GORM supports both, production-grade + zero-config local dev |
| **IaC Export** | **Pulumi-first, Terraform export** | Pulumi as primary, export to Terraform for ecosystem compatibility |

---

## Background

Multiple BytePort workstreams exist across different locations that need consolidation:

### Sources

| Source | Location | Scope | Date |
|--------|----------|-------|------|
| **Archive** | `/Users/kooshapari/CodeProjects/archive/Rust/webApp/byte_port` | Hexagonal architecture (70%), 8 providers, REARCHITECTURE | Oct 2025 |
| **Repos (canonical)** | `/Users/kooshapari/CodeProjects/Phenotype/repos/BytePort` | Clean clone, PRD/FR/ADR docs | Mar 2026 |

### Related Projects

| Project | Location | Relationship |
|---------|----------|--------------|
| **nvms** | `/Users/kooshapari/CodeProjects/Phenotype/repos/nanovms` | Compute substrate - BytePort sits above it |
| **phenotype-infrakit** | `/Users/kooshapari/CodeProjects/Phenotype/repos` | Infrastructure crate library |

---

## Architecture Decisions

### 1. nvms Integration: **Dual Approach**

```
┌─────────────────────────────────────────────────────────────┐
│                      BytePort                                │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐    │
│  │            nvms Client Library (Go)                  │    │
│  │  • Direct sandbox lifecycle                         │    │
│  │  • Same-process execution                           │    │
│  │  • Shared memory space                              │    │
│  └─────────────────────────────────────────────────────┘    │
│                           │                                  │
│                           │ IPC (gRPC/Unix Socket)          │
│                           ▼                                  │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              nvms CLI Daemon                        │    │
│  │  • Cross-platform wrapper                           │    │
│  │  • Sandbox isolation (separate process)             │    │
│  │  • For non-Go SDKs (Python, TypeScript)            │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

**Why Dual?**
- **Library**: Performance, same-process control, transactional operations
- **CLI/IPC**: Multi-platform support, language-agnostic, process isolation
- **Pattern**: Go backend uses library; SDKs use CLI over IPC

### 2. Infrastructure as Code: **Pulumi**

```
┌─────────────────────────────────────────────────────────────┐
│                    Pulumi Stack                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐   │
│  │   Python    │  │    Go       │  │   TypeScript    │   │
│  │   SDK       │  │    SDK      │  │   SDK           │   │
│  └──────┬──────┘  └──────┬──────┘  └────────┬────────┘   │
│         │                  │                    │            │
│         └──────────────────┼────────────────────┘            │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Pulumi Engine                            │    │
│  │  • State management                                  │    │
│  │  • Resource graph                                   │    │
│  │  • Diff & planning                                  │    │
│  └─────────────────────────────────────────────────────┘    │
│                            │                                 │
│         ┌──────────────────┼──────────────────┐             │
│         ▼                  ▼                  ▼              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐     │
│  │    AWS      │  │    GCP      │  │     Azure       │     │
│  │   Provider  │  │   Provider  │  │    Provider     │     │
│  └─────────────┘  └─────────────┘  └─────────────────┘     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Why Pulumi over Terraform/SST?**
- **vs Terraform**: Full programming language (conditionals, loops, functions), better testing (unit/integration), superior debugging
- **vs SST v3**: Pulumi is cloud-agnostic, SST is AWS-first; BytePort targets multi-cloud
- **Maturity**: 7+ years, production-proven, excellent Go/Python/TS SDKs

### 3. Database: **PostgreSQL (Prod) + SQLite (Local)**

```
┌─────────────────────────────────────────────────────────────┐
│                    GORM Adapter Layer                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              BytePort Domain Models                  │    │
│  │  • User, Project, Deployment, Provider               │    │
│  │  • DeploymentLog, Credential, Secret                  │    │
│  └─────────────────────────────────────────────────────┘    │
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              GORM (Database Agnostic)                 │    │
│  └─────────────────────────────────────────────────────┘    │
│                            │                                 │
│         ┌──────────────────┼──────────────────┐             │
│         ▼                  ▼                  ▼              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐     │
│  │ PostgreSQL  │  │  SQLite     │  │   (Extensible)   │     │
│  │  (Prod)     │  │  (Local)    │  │   MySQL/PG      │     │
│  └─────────────┘  └─────────────┘  └─────────────────┘     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Why Both?**
- **PostgreSQL**: ACID transactions, concurrent connections, JSONB for metadata, production-grade
- **SQLite**: Zero-config, portable, perfect for local dev, CI/CD testing
- **GORM**: Already in codebase, handles both seamlessly

---

## Directory Structure

```
BytePort/
├── backend/
│   ├── api/                           # Go API (hexagonal architecture)
│   │   ├── internal/
│   │   │   ├── domain/               # Entities, value objects
│   │   │   ├── application/          # Use cases, DTOs
│   │   │   └── infrastructure/       # Adapters
│   │   │       ├── persistence/      # GORM repositories
│   │   │       ├── http/             # Gin handlers
│   │   │       ├── nvms/             # nvms client library
│   │   │       └── providers/         # Cloud provider adapters
│   │   └── main.go
│   ├── providers/                    # Provider implementations
│   │   ├── aws/
│   │   ├── azure/
│   │   ├── gcp/
│   │   ├── vercel/
│   │   ├── netlify/
│   │   ├── railway/
│   │   ├── flyio/
│   │   └── supabase/
│   └── pulumi/                       # Pulumi IaC stacks
│       ├── stacks/                   # Per-environment stacks
│       │   ├── dev/
│       │   ├── staging/
│       │   └── prod/
│       └── components/               # Reusable Pulumi components
│
├── frontend/
│   └── web-next/                     # Next.js dashboard
│
├── nvms-integration/                 # BytePort ↔ nvms bridge
│   ├── client/                       # Go library (same-process)
│   │   └── nvms.go
│   ├── ipc/                          # IPC layer (CLI daemon)
│   │   ├── socket.go
│   │   └── proto/
│   └── sdk/                          # Multi-language SDKs
│       ├── python/
│       └── typescript/
│
├── manifests/                        # IaC manifest schemas
│   ├── schema.yaml                   # JSON Schema
│   └── examples/
│
├── byteport.py                       # Orchestrator (KInfra-powered)
│
├── sdk/
│   ├── go/
│   ├── python/
│   └── typescript/
│
└── docs/
    ├── REARCHITECTURE.md             # Full design doc (preserve)
    ├── PROVIDERS.md                  # Provider implementation details
    └── ARCHITECTURE.md              # Hexagonal architecture
```

---

## Merger Strategy

### Step 1: Copy Archive to Canonical

```bash
# Source: Oct 2025 work
cp -r /archive/Rust/webApp/byte_port/* /repos/BytePort/

# Preserve these specific files:
# - REARCHITECTURE.md
# - PHASE_2_COMPLETE.md
# - All hexagonal architecture code
# - All 8 provider implementations
```

### Step 2: Add Missing Provider Files

3 untracked files from archive:
- `backend/api/lib/cloud/provider_netlify.go`
- `backend/api/lib/cloud/provider_railway.go`
- `backend/api/lib/cloud/provider_vercel.go`

### Step 3: Add nvms Integration

Create `backend/api/internal/infrastructure/nvms/`:
- `client.go` - Library interface
- `ipc.go` - CLI/IPC client
- `sandbox.go` - Sandbox lifecycle management

### Step 4: Add Pulumi Infrastructure

Create `backend/pulumi/`:
- `stacks/dev/Pulumi.yaml`
- `stacks/prod/Pulumi.yaml`
- `components/ecs_service.go` - Reusable ECS component
- `components/lambda_function.go` - Serverless component

### Step 5: Database Configuration

Update `backend/api/internal/infrastructure/persistence/`:
- Add SQLite support for local dev
- Keep PostgreSQL for production
- Auto-detect based on `BP_ENV=local|production`

---

## PRD Update (Expanded Scope)

```markdown
## Epics (v2)

### E1 — IaC Manifest Parsing
- [x] Manifest schema (YAML/JSON)
- [x] Multi-service support
- [x] Provider-agnostic spec
- [ ] nvms integration layer
- [ ] Cost optimization hints

### E2 — Multi-Cloud Deployment
- [x] AWS (EC2/ECS/Lambda)
- [x] Azure
- [x] GCP
- [x] Vercel, Netlify, Railway, Fly.io, Supabase
- [x] Pulumi infrastructure as code
- [ ] SST v3 patterns (optional)

### E3 — Compute Targets
- [ ] Local: vLLM/MLX + nvms + Docker
- [ ] SaaS: Vercel/Supabase/Railway (freemium)
- [ ] Cloud: AWS/GCP/Azure + Pulumi
- [ ] nvms as deployment target

### E4 — Portfolio UX Generation
- [ ] LLM-powered template enhancement
- [ ] Interactive deployment widgets
- [ ] Real-time status embedding

### E5 — nvms Substrate
- [ ] BytePort ↔ nvms integration
- [ ] VM/sandbox lifecycle APIs
- [ ] Multi-tier isolation support
```

---

## Implementation Phases

### Phase 1: Consolidation ✅
- [x] Plan created
- [ ] Copy archive to canonical repo
- [ ] Add 3 missing provider files
- [ ] Verify hexagonal architecture builds
- [ ] Verify all providers compile

### Phase 2: nvms Integration
- [ ] Create `nvms-integration/` module
- [ ] Implement Go client library
- [ ] Implement IPC layer
- [ ] Add Python/TS SDK wrappers

### Phase 3: Pulumi Integration
- [ ] Create Pulumi stack structure
- [ ] Add AWS ECS component
- [ ] Add Lambda component
- [ ] Add multi-cloud support

### Phase 4: Database Polish
- [ ] Add SQLite adapter
- [ ] Auto-detect environment
- [ ] Update connection config

### Phase 5: Frontend Completion
- [ ] Complete Phase 3.4-3.6 (frontend consolidation)
- [ ] Add provider comparison UI
- [ ] Add cost optimizer dashboard

### Phase 6: Testing & Docs
- [ ] Integration tests
- [ ] E2E tests for providers
- [ ] Update documentation

---

## Key Files to Preserve

From archive Oct 2025 work:

| File | Purpose |
|------|---------|
| `REARCHITECTURE.md` | Full design doc with rationale |
| `PHASE_2_COMPLETE.md` | Provider implementation details |
| `backend/api/internal/domain/` | Hexagonal domain layer |
| `backend/api/internal/application/` | Hexagonal application layer |
| `backend/api/internal/infrastructure/` | Hexagonal infrastructure layer |
| `backend/api/lib/cloud/` | All 8 provider implementations |

---

## Next Steps

1. Execute: Copy archive to canonical repo
2. Add the 3 untracked provider files
3. Verify build compiles
4. Create nvms integration module
5. Add Pulumi infrastructure stacks
