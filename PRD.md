# PRD — BytePort

## Overview

BytePort is a multi-cloud Infrastructure-as-Code (IaC) deployment platform with portfolio UX generation. Developers define their application and infrastructure in a single manifest; BytePort deploys to any cloud provider (local nvms, AWS, GCP, Azure, or freemium SaaS) and generates interactive portfolio site templates to showcase deployed projects.

**BytePort sits above nvms** as one of its compute targets, alongside traditional cloud providers.

## Core Vision

A unified deployment platform that:
1. **Abstracts cloud complexity** — One manifest, any target
2. **Enables local-first** — Self-host with nvms at zero cost
3. **Scales to production** — Enterprise cloud when ready
4. **Generates portfolio** — Automatic project showcase

## Deployment Tiers

### Tier 1: Local ($0/month)
- **Compute**: nvms (self-hosted hypervisor on bare metal)
- **LLM**: vLLM/MLX (OpenAI-compatible API, self-hosted)
- **Storage**: Local filesystem / external NAS
- **Use**: Development, personal projects, privacy-focused

### Tier 2: Freemium SaaS ($0-10/month)
- **Frontend**: Vercel (100GB bandwidth free)
- **Backend**: Railway (500 hours/month free)
- **Database**: Supabase (500MB database, 1GB files)
- **LLM**: Ollama on Railway or cloud LLM API
- **Use**: Side projects, MVPs, small apps

### Tier 3: Production Cloud ($$$)
- **Compute**: AWS EC2/GKE, GCP GCE/GKE, Azure VMs/AKS
- **Storage**: S3/GCS/Blob Storage
- **Database**: RDS/Cloud SQL/SQL Database
- **CDN**: CloudFront/Cloud CDN/Azure CDN
- **LLM**: OpenAI/Anthropic/Vertex AI
- **IaC**: Pulumi (primary), Terraform (export)
- **Use**: Production applications, scaled services

## Cloud Providers

### Supported Providers

| Provider | Type | Tier | Status |
|----------|------|------|--------|
| **nvms** | Hypervisor | Local | Primary |
| **AWS** | Cloud | Production | Complete |
| **GCP** | Cloud | Production | Complete |
| **Azure** | Cloud | Production | Complete |
| **Vercel** | SaaS | Freemium | Complete |
| **Railway** | SaaS | Freemium | Complete |
| **Netlify** | SaaS | Freemium | Complete |
| **Fly.io** | SaaS | Freemium | Planned |
| **Supabase** | SaaS | Freemium | Planned |

## Epics

### E1 — IaC Manifest Parsing

| Story | Description | Acceptance Criteria |
|-------|-------------|---------------------|
| E1.1 | Developer writes BytePort manifest describing app + infra | Manifest parses; schema validated |
| E1.2 | Manifest supports multi-service applications | Each service = distinct deployable unit |
| E1.3 | Manifest supports all cloud targets | Provider-agnostic schema |

### E2 — Multi-Cloud Deployment

| Story | Description | Acceptance Criteria |
|-------|-------------|---------------------|
| E2.1 | Deploy to nvms (local hypervisor) | VM/container runs on local hardware |
| E2.2 | Deploy to AWS/GCP/Azure | Resources appear in cloud console |
| E2.3 | Deploy to SaaS (Vercel/Railway) | App accessible via provider URL |
| E2.4 | Pull source from GitHub | Correct branch/ref deployed |
| E2.5 | Deployment status reporting | Live URLs on success |

### E3 — LLM Integration

| Story | Description | Acceptance Criteria |
|-------|-------------|---------------------|
| E3.1 | Connect to vLLM/MLX (local) | OpenAI-compatible API |
| E3.2 | Connect to OpenAI/Anthropic | Cloud LLM API |
| E3.3 | AI-assisted manifest generation | LLM helps write IaC |

### E4 — Portfolio UX Generation

| Story | Description | Acceptance Criteria |
|-------|-------------|---------------------|
| E4.1 | Generate portfolio components | Object templates emitted |
| E4.2 | Embed live project endpoints | Portfolio widgets with live data |
| E4.3 | AI-enhanced descriptions | LLM generates showcase text |

### E5 — CLI Interface

| Story | Description | Acceptance Criteria |
|-------|-------------|---------------------|
| E5.1 | `byteport deploy` — full pipeline | Deploy + portfolio generation |
| E5.2 | `byteport status` — check health | Per-service health + URLs |
| E5.3 | `byteport init` — scaffold project | Generate manifest template |
| E5.4 | `byteport targets` — list providers | Show available deploy targets |

### E6 — Infrastructure as Code (Pulumi)

| Story | Description | Acceptance Criteria |
|-------|-------------|---------------------|
| E6.1 | Pulumi program structure | Reusable components |
| E6.2 | Export to Terraform | Compatibility layer |
| E6.3 | Multi-environment stacks | dev/staging/prod |

## Architecture

```
BytePort CLI
    │
    ▼
Manifest Parser (YAML/TOML/JSON)
    │
    ▼
Deployment Engine
├── nvms (local hypervisor)
├── AWS / GCP / Azure
└── Vercel / Railway / Netlify
    │
    ▼
Portfolio Generator (LLM-powered)
```

## Technical Stack

| Component | Technology |
|-----------|------------|
| Backend | Go (hexagonal architecture) |
| Frontend | Next.js 14 (App Router) |
| CLI | Go + Cobra |
| IaC | Pulumi (Go SDK) |
| Database | PostgreSQL (production), SQLite (local) |
| LLM | vLLM/MLX (local), OpenAI (cloud) |
| Auth | WorkOS / Custom |

## Non-Goals

- Custom domain management (v1)
- Billing/cost management UI (v1)
- Kubernetes management (v1)
- Real-time collaboration (v1)

## Roadmap

### Phase 1: Foundation
- [x] Go hexagonal backend structure
- [x] Cloud provider abstraction layer
- [x] AWS + GCP + Azure providers
- [x] Vercel + Railway + Netlify providers
- [x] nvms integration scaffold
- [x] Pulumi infrastructure scaffold

### Phase 2: Manifest & Deployment
- [ ] Manifest schema (YAML/TOML)
- [ ] Manifest parser + validator
- [ ] Multi-target deployment
- [ ] Deployment status tracking

### Phase 3: LLM Integration
- [ ] vLLM/MLX connector
- [ ] OpenAI connector
- [ ] AI-assisted manifest generation

### Phase 4: Portfolio
- [ ] Portfolio template generator
- [ ] Component library
- [ ] Live endpoint embedding

### Phase 5: CLI
- [ ] `byteport deploy`
- [ ] `byteport status`
- [ ] `byteport init`
- [ ] `byteport targets`

---

*Last Updated: 2026-04-02*
