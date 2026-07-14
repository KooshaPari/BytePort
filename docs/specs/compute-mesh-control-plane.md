# BytePort compute-mesh control plane

BytePort models the organization as one Kubernetes-like compute mesh spanning
cloud providers, local hosts, devices, and managed services. A provider is an
adapter behind this control plane, not a separate source of truth.

## Mesh vocabulary

- **Node**: a host, VM, device, cluster, or managed service endpoint with stable
  identity and capabilities.
- **Capability**: an advertised resource or feature (CPU, memory, GPU, region,
  network, database, queue, container backend, or policy).
- **Workload**: a PhenoCompose artifact or provider-native resource with desired
  placement, dependencies, health, and lifecycle.
- **Provider adapter**: an implementation for Vercel, Supabase, Neon, Upstash,
  GCP, AWS, Azure, Hetzner, Netlify, Render, and future providers.
- **Execution adapter**: NanoVMS, Podman, Apple Containers, or WSL Containers;
  it reports runtime status but does not own desired state.

The control plane owns desired state, placement decisions, reconciliation status,
and audit evidence. PhenoCompose owns composition validation/rendering. NanoVMS
owns execution-engine selection and lifecycle. substrate/sharecli/phenodag may
submit intents and observe status, but cannot create competing state stores.

Every workload handoff carries an immutable composition digest, artifact reference,
owner, source, verification result, and evidence locator. Provider credentials stay
inside the provider adapter boundary.
