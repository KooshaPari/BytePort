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

The `POST /api/v1/mesh/workloads` contract requires the upstream handoff to include
`source` and `evidence`; `owner` is taken from the authenticated principal and can
never be supplied by the request body. BytePort validates the digest and references
before accepting the intent, returns `verified: true`, and persists the trace fields
alongside the composition metadata. A replay with the same owner, name, digest, and
trace fields returns the original workload identity. A changed digest or trace
locator is a conflict, never an implicit replacement.

```json
{
  "name": "phenotype-lab",
  "composition_digest": "sha256:<64-hex-digits>",
  "artifact_ref": "oci://registry/example",
  "execution_backend": "podman",
  "source": "git://github.com/KooshaPari/PhenoCompose/examples/composition-v0.yaml",
  "evidence": "run://phenocompose/<receipt>"
}
```
