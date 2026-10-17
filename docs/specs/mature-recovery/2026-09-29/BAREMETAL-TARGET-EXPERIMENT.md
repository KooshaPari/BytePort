# BytePort bare-metal target bootstrap experiment — v0.1

Date: 2026-09-30.
Status: **B09 DESIGN EXPERIMENT / no provider winner frozen**.

## Objective

Compare the leading existing bare-metal engines behind BytePort's generalized TargetAdapter contract before implementing low-level hardware lifecycle ourselves.

Candidates:
1. Ironic standalone, bootstrapped with Bifrost;
2. MAAS;
3. Tinkerbell.

## Required disposable environment

Preferred first environment:
- Linux host with KVM/libvirt;
- >=2 disposable VMs representing managed machines;
- isolated test network where DHCP/PXE manipulation is safe;
- no production BMC credentials.

If a candidate cannot be realistically exercised with virtual BMC/test nodes, record that setup limitation rather than replacing the experiment with mocks.

## Common lifecycle assignment

For each candidate:

1. install/control-plane bootstrap;
2. register/enroll a target;
3. inspect/commission/discover hardware identity;
4. obtain stable external resource identity;
5. provision a known Linux image;
6. observe installed/running state;
7. restart the provider control plane;
8. rediscover the same target/resource;
9. adopt or reconcile pre-existing realized state where supported;
10. reprovision/update a controlled property;
11. power/lifecycle operation;
12. deprovision/release/clean;
13. verify unrelated target survives;
14. remove provider inventory record separately from wiping/deprovisioning.

## BytePort adapter mapping

Record how candidate primitives map to:
- Target;
- TargetCapabilities;
- DesiredResource(host);
- ExternalOperationRef;
- RealizedResource;
- InfrastructureObservation;
- CREATE / UPDATE / REPLACE / READ / DELETE / UNKNOWN;
- adoption;
- DestructionIntent.

Any semantic mismatch is recorded explicitly.

## Measurements

### Bootstrap/operating burden
- install steps;
- privileged/system dependencies;
- services/containers/VMs;
- memory/disk footprint;
- time to ready;
- network assumptions;
- Kubernetes/OpenStack dependency footprint.

### Lifecycle coverage
- discovery/inspection;
- adoption;
- BMC/power;
- image provisioning;
- storage/network configuration;
- cleanup/wipe;
- rescue;
- event/operation tracking;
- restart recovery.

### Integration quality
- API quality/stability;
- external operation identity;
- idempotency;
- observation freshness;
- failure taxonomy;
- auth/multi-user boundary;
- client/library availability;
- provider-specific leakage into BytePort domain.

### Project viability
- licensing;
- active maintenance;
- release cadence;
- documentation;
- security posture.

## Hard rejects

A candidate is not acceptable as the primary bootstrap if:
- BytePort must reimplement most BMC/PXE/imaging behavior around it;
- stable target/resource identity cannot be recovered after restart;
- destructive lifecycle states cannot be distinguished;
- integration requires making Kubernetes/OpenStack semantics BytePort's product identity;
- test/operational burden exceeds realistic value versus another candidate.

## Expected outcome

The experiment may choose different adapters for different environments. It does not require one global winner.

Current hypothesis:
- Ironic+Bifrost leads low-level lifecycle breadth/minimal hand-rolling;
- MAAS may lead integrated inventory/operator lifecycle;
- Tinkerbell may lead customizable provisioning workflows.

No hypothesis is a gate result until this assignment executes.
