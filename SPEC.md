# BytePort — SPEC.md

## Overview

BytePort is an Infrastructure-as-Code (IaC) deployment platform combined with portfolio UX generation. Developers define applications and AWS infrastructure in a single NVMS manifest; BytePort deploys to AWS and automatically generates portfolio site components showcasing deployed projects.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         BytePort                                 │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Management UI                          │  │
│  │  (Frontend - Web Interface for Deployment Management)     │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌──────────────────┐    │  │
│  │  │   Dashboard │ │  Deploy     │ │  Portfolio       │    │  │
│  │  │   (Status)  │ │  Wizard     │ │  Preview       │    │  │
│  │  └─────────────┘  └─────────────┘  └──────────────────┘    │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌───────────────────────────┴──────────────────────────────┐  │
│  │                  ByteBridge API (Go)                      │  │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐           │  │
│  │  │  Manifest  │ │   Deploy    │ │  Portfolio │           │  │
│  │  │  Parser    │ │  Engine     │ │  Generator │           │  │
│  │  │            │ │            │ │            │           │  │
│  │  │ • Validate │ │ • AWS SDK   │ │ • LLM      │           │  │
│  │  │ • Transform│ │ • NanoVMS   │ │ • Templates│           │  │
│  │  └────────────┘ └────────────┘ └────────────┘           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌───────────────────────────┴──────────────────────────────┐  │
│  │                   NanoVMS Layer                           │  │
│  │  ┌──────────────────────────────────────────────────┐      │  │
│  │  │           MicroVM Orchestration                  │      │  │
│  │  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐  │      │  │
│  │  │  │ VM 1   │ │ VM 2   │ │ VM 3   │ │ VM N   │  │      │  │
│  │  │  │(App A) │ │(App B) │ │(App C) │ │(App N) │  │      │  │
│  │  │  └────────┘ └────────┘ └────────┘ └────────┘  │      │  │
│  │  │           Lightweight, isolated VMs            │      │  │
│  │  └──────────────────────────────────────────────────┘      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌───────────────────────────┴──────────────────────────────┐  │
│  │                    AWS Infrastructure                       │  │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐           │  │
│  │  │    EC2     │ │    ECS     │ │   Lambda   │           │  │
│  │  │  (Compute) │ │(Containers)│ │ (Serverless)│          │  │
│  │  └────────────┘ └────────────┘ └────────────┘           │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Components

### Core Components

| Component | Responsibility | Interface |
|-----------|----------------|-----------|
| `ManifestParser` | NVMS manifest validation | `Parse(manifest []byte) (*DeploymentConfig, error)` |
| `DeployEngine` | AWS resource provisioning | `Deploy(config *DeploymentConfig) (*Deployment, error)` |
| `NanoVMSManager` | MicroVM lifecycle | `CreateVM(spec VMSpec) (*VM, error)` |
| `PortfolioGenerator` | UX template generation | `Generate(project *Project) (*PortfolioPage, error)` |
| `LLMBackend` | Template text generation | `GenerateDescription(project Project) (string, error)` |

### NVMS Manifest Format

```yaml
# odin.nvms - Example manifest
NAME: my-app
DESCRIPTION: A web application for task management

SERVICES:
  - NAME: "main"
    PATH: "./frontend"
    PORT: 8080
    ENV:
      - API_URL=http://localhost:8081
      
  - NAME: "backend"
    PATH: "./backend"
    PORT: 8081
    ENV:
      - DATABASE_URL=postgres://localhost/mydb

INFRASTRUCTURE:
  compute: ec2        # or ecs, lambda
  region: us-east-1
  instance_type: t3.micro

PORTFOLIO:
  generate_page: true
  screenshots: auto   # auto-capture on deploy
  description_source: llm  # or readme, manual
```

---

## Data Models

### Deployment Configuration

```go
type DeploymentConfig struct {
    Name        string            `yaml:"NAME"`
    Description string            `yaml:"DESCRIPTION"`
    Services    []Service         `yaml:"SERVICES"`
    Infra       Infrastructure    `yaml:"INFRASTRUCTURE"`
    Portfolio   PortfolioConfig   `yaml:"PORTFOLIO"`
}

type Service struct {
    Name string            `yaml:"NAME"`
    Path string            `yaml:"PATH"`
    Port int               `yaml:"PORT"`
    Env  map[string]string `yaml:"ENV"`
}

type Infrastructure struct {
    Compute      string `yaml:"compute"`
    Region       string `yaml:"region"`
    InstanceType string `yaml:"instance_type"`
}
```

### Deployment State

```go
type Deployment struct {
    ID          string
    Name        string
    Status      DeploymentStatus
    Services    []ServiceStatus
    VM          *VMInfo
    URL         string
    PortfolioURL string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type DeploymentStatus string
const (
    StatusPending     DeploymentStatus = "PENDING"
    StatusPreparing   DeploymentStatus = "PREPARING"
    StatusDeploying   DeploymentStatus = "DEPLOYING"
    StatusRunning     DeploymentStatus = "RUNNING"
    StatusFailed      DeploymentStatus = "FAILED"
    StatusTerminated  DeploymentStatus = "TERMINATED"
)
```

### Portfolio Page

```go
type PortfolioPage struct {
    ProjectID   string
    Title       string
    Description string
    ScreenshotURLs []string
    DeployURL   string
    RepoURL     string
    TechStack   []string
    GeneratedAt time.Time
}
```

---

## Stack

| Layer | Technology | Notes |
|-------|-----------|-------|
| Backend | Go | Deployment engine |
| Frontend | Web (vanilla/Svelte) | Management UI |
| IaC Format | NVMS | Custom YAML-based |
| Cloud | AWS | EC2 primary, ECS/Lambda optional |
| Virtualization | NanoVMS | Custom MicroVM platform |
| LLM | OpenAI / LLaMA | Template generation |
| CLI | SpinCLI | VM management tool |

---

## API Contract

### Deploy Application

```
POST /api/v1/deployments
Content-Type: application/json

Request:
{
  "repo_url": "https://github.com/user/repo",
  "branch": "main",
  "manifest_path": "odin.nvms",
  "name": "my-app"
}

Response: 202 Accepted
{
  "deployment_id": "dep_abc123",
  "status": "PENDING",
  "stream_url": "/api/v1/deployments/dep_abc123/stream"
}
```

### Get Deployment Status

```
GET /api/v1/deployments/:id

Response: 200 OK
{
  "id": "dep_abc123",
  "name": "my-app",
  "status": "RUNNING",
  "services": [
    {"name": "main", "status": "healthy", "url": "http://..."},
    {"name": "backend", "status": "healthy", "url": "http://..."}
  ],
  "vm": {
    "id": "vm_xyz789",
    "ip": "3.91.42.100",
    "status": "running"
  },
  "portfolio_url": "https://portfolio.example.com/projects/my-app"
}
```

### Generate Portfolio Page

```
POST /api/v1/portfolio/generate

Request:
{
  "deployment_id": "dep_abc123",
  "options": {
    "include_screenshots": true,
    "description_source": "llm",
    "template": "modern"
  }
}

Response: 200 OK
{
  "page_url": "https://portfolio.example.com/projects/my-app",
  "generated_description": "A task management application built with...",
  "screenshots": [
    "https://cdn.example.com/screenshots/my-app-1.png"
  ]
}
```

---

## NanoVMS Integration

| Feature | Implementation |
|---------|----------------|
| VM Creation | SpinCLI + Firecracker |
| Image Building | Dockerfile → MicroVM image |
| Networking | VPC + Security groups |
| Storage | EBS volumes per VM |
| Scaling | Horizontal VM pools |

### VM Lifecycle

```
Create → Configure → Start → Health Check → Register
  ↓        ↓         ↓          ↓              ↓
  └────────┴─────────┴──────────┴──────────────┘
                         ↓
                      Running
                         ↓
      ┌──────────────────┼──────────────────┐
      ↓                  ↓                  ↓
   Update            Terminate          Snapshot
```

---

## Performance

| Metric | Target | Measurement |
|--------|--------|-------------|
| Deploy latency | <5 min | Repo pull to running |
| VM cold start | <2s | NanoVMS boot time |
| Portfolio generation | <30s | LLM + screenshots |
| Concurrent deploys | 10+ | Per BytePort instance |
| Uptime SLA | 99.9% | Deployed applications |

---

## Project Structure

```
BytePort/
├── backend/
│   ├── byteport/             # Core deployment engine
│   │   ├── cmd/              # CLI commands
│   │   ├── pkg/
│   │   │   ├── manifest/     # NVMS parser
│   │   │   ├── deploy/       # AWS deployment
│   │   │   ├── nanovms/      # VM management
│   │   │   └── portfolio/    # UX generation
│   │   └── main.go           # Entry point
│   └── bytebridge/           # Integration layer
│       ├── api/              # REST handlers
│       ├── middleware/       # Auth, logging
│       └── server.go         # HTTP server
├── frontend/                 # Management UI
│   ├── src/
│   └── public/
├── odin.nvms                 # Example manifest
└── start                     # Local dev script
```

---

## References

- [AWS SDK for Go](https://aws.github.io/aws-sdk-go-v2/)
- [Firecracker MicroVMs](https://firecracker-microvm.github.io/)
- [SpinCLI Documentation](https://developer.fermyon.com/spin)
- [NVMS Manifest Spec](./docs/NVMS_SPEC.md)
