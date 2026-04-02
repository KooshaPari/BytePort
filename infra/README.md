# BytePort Infrastructure — Pulumi

Infrastructure-as-Code using [Pulumi](https://www.pulumi.com/) for multi-cloud deployments.

## Structure

```
infra/
├── Pulumi.yaml           # Pulumi project config
├── Pulumi.dev.yaml       # Dev stack config (secrets)
├── Pulumi.staging.yaml   # Staging stack config
├── Pulumi.prod.yaml      # Production stack config
├── go/                   # Go Pulumi programs
│   ├── main.go           # Main entry point
│   ├── components/       # Reusable infrastructure components
│   │   ├── compute.go    # VM/container compute
│   │   ├── storage.go    # Object/block storage
│   │   └── network.go    # VPC/networking
│   └── targets/          # Provider-specific targets
│       ├── aws.go        # AWS resources
│       ├── gcp.go        # Google Cloud
│       ├── azure.go      # Azure
│       └── local.go      # Local/self-hosted (nvms)
└── scripts/              # Deployment scripts
    └── deploy.sh         # Deploy script
```

## Providers

### Supported Targets

| Target | Provider | Use Case |
|--------|----------|----------|
| **Local** | `local` | Development, self-hosted (nvms) |
| **Vercel** | API | Serverless frontend hosting |
| **Railway** | API | Containerized backend |
| **Netlify** | API | Static sites + functions |
| **Supabase** | Terraform/API | Database + auth |
| **AWS** | Pulumi AWS | Production workloads |
| **GCP** | Pulumi GCP | Production workloads |
| **Azure** | Pulumi Azure | Production workloads |

## Quick Start

### Prerequisites

```bash
# Install Pulumi
curl -fsSL https://get.pulumi.com | sh

# Or via brew
brew install pulumi

# Login to Pulumi
pulumi login

# Install Go dependencies
cd infra/go
go mod tidy
```

### Deploy to Local (nvms)

```bash
cd infra/go
pulumi up --stack dev --target=local
```

### Deploy to AWS

```bash
# Set AWS credentials
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret

# Deploy to AWS dev
pulumi up --stack dev --target=aws
```

## Deployment Tiers

### Tier 1: Local ($0/month)
- **Compute**: nvms (self-hosted hypervisor)
- **Storage**: Local filesystem / external NAS
- **Network**: LAN / tailscale VPN
- **Use**: Development, personal projects

### Tier 2: Freemium SaaS ($0-10/month)
- **Frontend**: Vercel (100GB bandwidth)
- **Backend**: Railway (500 hours/month)
- **Database**: Supabase (500MB database, 1GB files)
- **Use**: Side projects, MVPs, small apps

### Tier 3: Production Cloud ($$$)
- **Compute**: AWS EC2/GKE, GCP GCE/GKE, Azure VMs/AKS
- **Storage**: S3/GCS/Blob Storage
- **Database**: RDS/Cloud SQL/SQL Database
- **Network**: VPC, Load Balancer, CDN
- **Use**: Production applications, scaled services

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      BytePort API                            │
├─────────────┬─────────────┬─────────────┬──────────────────┤
│   Vercel    │  Railway    │   Netlify   │   AWS/GCP/Azure  │
│  (Frontend) │  (Backend)  │  (Functions)│    (Enterprise)  │
├─────────────┴─────────────┴─────────────┴──────────────────┤
│                    nvms Integration Layer                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌────────────────┐ │
│  │  nvms   │  │   AWS   │  │   GCP   │  │     Azure      │ │
│  │(Hypervis│  │   EC2   │  │   GCE   │  │      VMs       │ │
│  └─────────┘  └─────────┘  └─────────┘  └────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Configuration

### Stack Config

```yaml
# Pulumi.dev.yaml
config:
  byteport:environment: dev
  byteport:computeTier: freemium
  byteport:providers:
    - vercel
    - railway
    - supabase
```

## Notes

- Pulumi state is stored in Pulumi Cloud by default
- For self-hosted state, use `pulumi login --local` or configure a backend
- Secrets are encrypted in state files
- Each provider's SDK is imported only when that provider is targeted
