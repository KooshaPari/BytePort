# BytePort artifact/provenance evidence conventions — v1.0

Date: 2026-10-01.

## Artifact identity

For OCI outputs, canonical immutable identity is an OCI descriptor/manifest digest over exact content. Multi-platform outputs additionally preserve image-index identity and the exact selected platform manifest at runtime.

A local Docker image ID or exported archive SHA may be useful experimental evidence but does not substitute for canonical OCI registry/content identity when OCI distribution is the subject.

## Provenance

“Provenance complete” requires verifiable build provenance, not merely:
- builder name/version;
- CI job URL;
- source SHA in logs;
- artifact digest alone.

Preferred interoperable representation: SLSA/in-toto-compatible build provenance or an equivalently verifiable attestation binding:
- artifact subject digest;
- builder identity;
- build definition;
- external parameters/source;
- invocation/run;
- relevant materials;
- timestamps/metadata;
- signature/trust verification where policy requires it.

## Evidence states

- ARTIFACT_IDENTIFIED: immutable subject digest established.
- BUILD_RECEIPT: build execution metadata exists but is not attested provenance.
- PROVENANCE_ATTESTED: provenance predicate binds exact subject.
- PROVENANCE_VERIFIED: attestation verified under accepted trust policy.

Do not collapse these states.

## Runtime selection

For multi-platform OCI:
image index digest != necessarily the exact platform manifest digest executed. RuntimeAdapter evidence records both where applicable.

## Finality

This convention closes the non-code meaning of B04 provenance. The remaining blocker is implementation/verification of an attestation path, not further definition.
