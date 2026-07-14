# Execution substrate boundary

BytePort may accept execution-substrate metadata as an opaque deployment hint:

- `nanovms`
- `podman`
- `apple-containers`
- `wsl-containers`

The value is descriptive metadata only. It must never select a cloud provider,
change authenticated ownership, or cause BytePort to persist runtime credentials.
PhenoCompose validates and renders the plan; NanoVMS selects and supervises the
runtime adapter. Podman, Apple Containers, and the first-party WSL Containers
extension consume Docker-compatible OCI plans. NanoVMS consumes NanoVMS plans.

BytePort is the abstract deployment and compute-mesh control plane. It may hold
desired state and lifecycle records for Vercel, Supabase, Neon, Upstash, GCP, AWS,
Azure, Hetzner, Netlify, Render, and additional providers, plus devices and
services represented in the organization compute mesh. This is intentionally
Kubernetes-like at the inventory/scheduling boundary: nodes, capabilities,
workloads, dependencies, health, and lifecycle are represented uniformly while
provider adapters remain replaceable.

PhenoCompose still owns composition validation/rendering, and NanoVMS owns runtime
selection/lifecycle. All substrate handoffs must carry a verified composition
digest and artifact reference; runtime adapters never become BytePort state owners.
