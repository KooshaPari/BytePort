package models

import (
	"context"
	"fmt"
)

// NanoVMSTransport is the narrow transport boundary needed to adapt the
// historical NanoVMS runtime into the generalized infrastructure lifecycle.
// Production HTTP wiring is intentionally outside this model package.
type NanoVMSTransport interface {
	Deploy(context.Context, NanoVMSDeployRequest) (NanoVMSSandbox, error)
	Stop(context.Context, string) error
	Observe(context.Context, string) (NanoVMSSandbox, bool, error)
}

type NanoVMSDeployRequest struct {
	Name          string
	Image         string
	OperationID   string
	ResourceID    string
	ConfigDigest  string
}

type NanoVMSSandbox struct {
	ID     string
	Name   string
	Status string
}

// NanoVMSInfrastructureAdapter is a candidate production-adapter boundary.
// It is not wired into /deploy yet: artifact authority and production
// destructive authorization remain separate gates.
type NanoVMSInfrastructureAdapter struct {
	Transport NanoVMSTransport
	TargetID  string
	Provider  string
}

func (a NanoVMSInfrastructureAdapter) Capabilities(
	_ context.Context,
	targetID string,
) (TargetCapabilities, error) {
	if targetID == "" || targetID != a.TargetID {
		return TargetCapabilities{}, fmt.Errorf("unknown NanoVMS target %q", targetID)
	}
	return TargetCapabilities{
		TargetID:        a.TargetID,
		Provider:        a.Provider,
		SupportsObserve: true,
		SupportsCreate:  true,
		SupportsUpdate:  false,
		SupportsReplace: false,
		SupportsDelete:  true,
	}, nil
}

func (a NanoVMSInfrastructureAdapter) Observe(
	ctx context.Context,
	realized RealizedResource,
) (InfrastructureObservation, error) {
	if realized.TargetID != a.TargetID {
		return InfrastructureObservation{}, fmt.Errorf(
			"realized target %q does not match NanoVMS target %q",
			realized.TargetID,
			a.TargetID,
		)
	}
	if realized.Provider != "" && realized.Provider != a.Provider {
		return InfrastructureObservation{}, fmt.Errorf(
			"realized provider %q does not match NanoVMS provider %q",
			realized.Provider,
			a.Provider,
		)
	}
	if realized.ExternalID == "" {
		return InfrastructureObservation{}, fmt.Errorf("NanoVMS realized resource has no sandbox ID")
	}

	sandbox, found, err := a.Transport.Observe(ctx, realized.ExternalID)
	if err != nil {
		return InfrastructureObservation{}, err
	}
	if !found {
		return InfrastructureObservation{
			RealizedResourceID: realized.ID,
			TargetID:           a.TargetID,
			Provider:           a.Provider,
			State:              "unknown",
			Fresh:              false,
		}, nil
	}
	if sandbox.ID != realized.ExternalID {
		return InfrastructureObservation{}, fmt.Errorf(
			"NanoVMS observation returned sandbox %q for requested %q",
			sandbox.ID,
			realized.ExternalID,
		)
	}
	return InfrastructureObservation{
		RealizedResourceID: realized.ID,
		TargetID:           a.TargetID,
		Provider:           a.Provider,
		State:              sandbox.Status,
		Fresh:              true,
	}, nil
}

func (a NanoVMSInfrastructureAdapter) Apply(
	ctx context.Context,
	action PlannedResourceAction,
	desired *DesiredResource,
	realized *RealizedResource,
) (InfrastructureApplyResult, error) {
	switch action.Action {
	case ReconcileCreate:
		if desired == nil || realized != nil {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS CREATE requires desired state only")
		}
		if desired.Target != a.TargetID {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS CREATE target mismatch")
		}
		if desired.Artifact == nil || desired.Artifact.Digest == "" {
			return InfrastructureApplyResult{}, fmt.Errorf(
				"NanoVMS CREATE requires an authorized immutable artifact",
			)
		}
		sandbox, err := a.Transport.Deploy(ctx, NanoVMSDeployRequest{
			Name:         desired.ID,
			Image:        desired.Artifact.Digest,
			OperationID:  action.DesiredResourceID,
			ResourceID:   desired.ID,
			ConfigDigest: desired.ConfigDigest,
		})
		if err != nil {
			return InfrastructureApplyResult{}, err
		}
		if sandbox.ID == "" {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS CREATE returned no sandbox ID")
		}
		out := RealizedResource{
			ID:                "nanovms:" + sandbox.ID,
			DesiredResourceID: desired.ID,
			TargetID:          a.TargetID,
			Provider:          a.Provider,
			ExternalID:        sandbox.ID,
		}
		return InfrastructureApplyResult{Realized: &out}, nil

	case ReconcileDelete:
		if realized == nil {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS DELETE requires exact realized state")
		}
		if realized.TargetID != a.TargetID || realized.Provider != a.Provider ||
			realized.ExternalID == "" {
			return InfrastructureApplyResult{}, fmt.Errorf(
				"NanoVMS DELETE requires exact provider/target/sandbox identity",
			)
		}
		if err := a.Transport.Stop(ctx, realized.ExternalID); err != nil {
			return InfrastructureApplyResult{}, err
		}
		return InfrastructureApplyResult{
			ExternalOperation: &ExternalOperationRef{
				Provider:   a.Provider,
				TargetID:   a.TargetID,
				ExternalID: realized.ExternalID,
				LookupKind: "nanovms-stop",
			},
		}, nil

	case ReconcileNoop, ReconcileRead, ReconcileUnknown:
		return InfrastructureApplyResult{}, fmt.Errorf(
			"NanoVMS adapter Apply cannot execute non-mutation action %q",
			action.Action,
		)

	case ReconcileUpdate, ReconcileReplace:
		return InfrastructureApplyResult{}, fmt.Errorf(
			"NanoVMS adapter does not claim %s capability",
			action.Action,
		)

	default:
		return InfrastructureApplyResult{}, fmt.Errorf("unsupported NanoVMS action %q", action.Action)
	}
}
