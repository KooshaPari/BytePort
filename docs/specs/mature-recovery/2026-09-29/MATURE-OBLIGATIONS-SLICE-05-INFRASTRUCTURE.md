# BytePort mature obligations — semantic slice 05: generalized infrastructure desired state

Status: **PROVISIONAL / DIRECT-USER-THESIS DERIVED**  
Date: 2026-09-30.

## Desired state

### BP-MO-INFRA-001
ManifestRevision resolves to a versioned DesiredResourceGraph, not only an application service list.

### BP-MO-INFRA-002
Each DesiredResource has stable identity, kind, configuration digest and dependency/relationship edges.

### BP-MO-INFRA-003
Runnable services/jobs, managed resources, hosts/machines, network/storage and adapter-defined resources may coexist without fake BuildArtifacts.

### BP-MO-INFRA-004
Target identity/configuration is explicit and independent from Project identity.

### BP-MO-INFRA-005
PlacementConstraint expresses target/region/host/capability/policy requirements without hard-coding one cloud provider into the domain.

## Planning / realization

### BP-MO-INFRA-010
ResourcePlan maps desired graph to adapter operations and build requirements but is not realized state.

### BP-MO-INFRA-011
Adapter capability mismatch fails explicitly before destructive approximation.

### BP-MO-INFRA-012
RealizedResource preserves exact target/provider identity and the DesiredResource/generation it realizes.

### BP-MO-INFRA-013
Application ProviderResource remains a compatible specialization/projection of RealizedResource.

## Observation / reconciliation

### BP-MO-INFRA-020
InfrastructureObservation records exact observed subject, state/config digest, freshness and provenance.

### BP-MO-INFRA-021
Reconciliation compares desired and observed graphs and classifies create/update/replace/delete/no-op/unknown.

### BP-MO-INFRA-022
Insufficient observation produces UNKNOWN rather than destructive convergence guesses.

### BP-MO-INFRA-023
Reconciliation survives control-plane restart through durable Operation/ExternalOperationRef/RealizedResource identity.

## Update / destruction

### BP-MO-INFRA-030
Changing desired state creates an explicit new revision/generation and traceable reconciliation plan.

### BP-MO-INFRA-031
Resource replacement preserves old/new identities and transition state until accepted observation.

### BP-MO-INFRA-032
Destruction requires explicit authorized DestructionIntent identifying exact desired/realized resources.

### BP-MO-INFRA-033
Removing BytePort locally is never implicit infrastructure destruction.

## Heterogeneous targets

### BP-MO-INFRA-040
Cloud and bare-metal adapters implement the same lifecycle contract where semantically applicable while declaring differing capabilities.

### BP-MO-INFRA-041
A bare-metal target can be represented without pretending it is a cloud account/region.

### BP-MO-INFRA-042
Provider-specific features may be exposed as capabilities/extensions without contaminating portable core identity.

## First generalized-infrastructure proof after selected-app slice

Extend the selected-app fixture with:
- one runnable service;
- one non-artifact managed/resource fixture;
- explicit Target;
- dependency edge;
- desired graph digest;
- realized-resource observations;
- update that changes one resource;
- reconciliation plan;
- explicit destruction intent;
- unrelated resource survival.

A later target-comparison slice should realize an equivalent portable subset on two materially different targets, with bare metal as a priority target family.
