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

BytePort remains the sole owner of IaC/cloud state, and all substrate handoffs must
carry a verified composition digest and artifact reference.
