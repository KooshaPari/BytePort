package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// NanoVMSTransport is the narrow transport boundary needed to adapt the
// historical NanoVMS runtime into the generalized infrastructure lifecycle.
// Production HTTP wiring is intentionally outside this model package.
type NanoVMSTransport interface {
	Deploy(context.Context, NanoVMSDeployRequest) (NanoVMSSandbox, error)
	Stop(context.Context, string) error
	Observe(context.Context, string) (NanoVMSSandbox, bool, error)
}

// NanoVMSDeletionTransport is separate from Stop. Implement it only when the
// provider contract supports actual resource removal and operation identity.
// The generic HTTP transport intentionally does not claim this capability.
// Acknowledgement alone is insufficient: Apply also observes the exact resource
// after Delete and remains UNKNOWN while it is present or observation fails.
type NanoVMSDeletionTransport interface {
	Delete(context.Context, string, string) error
}

func nanoVMSNonemptyIdentity(value string) bool {
	return strings.TrimSpace(value) != "" && strings.TrimSpace(value) == value &&
		!strings.ContainsAny(value, "\x00\r\n\t")
}

// NanoVMSOutcomeUnknownError means the transport cannot determine whether a
// mutation reached/committed at the provider. Callers must reconcile rather
// than blindly retry the mutation.
type NanoVMSOutcomeUnknownError struct {
	Operation string
	Cause     error
}

func (e *NanoVMSOutcomeUnknownError) Error() string {
	return fmt.Sprintf("NanoVMS %s outcome unknown: %v", e.Operation, e.Cause)
}

func (e *NanoVMSOutcomeUnknownError) Unwrap() error { return e.Cause }

type NanoVMSDeployRequest struct {
	Name         string
	Image        string
	OperationID  string
	ResourceID   string
	ConfigDigest string
}

type NanoVMSSandbox struct {
	ID           string
	Name         string
	Status       string
	ConfigDigest string
}

// NanoVMSInfrastructureAdapter is a candidate production-adapter boundary.
// It is not wired into /deploy yet: artifact authority and production
// destructive authorization remain separate gates.
type NanoVMSInfrastructureAdapter struct {
	Transport NanoVMSTransport
	Artifacts BuildArtifactResolver
	TargetID  string
	Provider  string
}

func (a NanoVMSInfrastructureAdapter) Capabilities(
	_ context.Context,
	targetID string,
) (TargetCapabilities, error) {
	if !nanoVMSNonemptyIdentity(a.TargetID) || !nanoVMSNonemptyIdentity(a.Provider) ||
		a.Transport == nil || targetID != a.TargetID {
		return TargetCapabilities{}, fmt.Errorf("unknown NanoVMS target %q", targetID)
	}
	_, canDelete := a.Transport.(NanoVMSDeletionTransport)
	return TargetCapabilities{
		TargetID:        a.TargetID,
		Provider:        a.Provider,
		SupportsObserve: true,
		SupportsCreate:  true,
		SupportsUpdate:  false,
		SupportsReplace: false,
		SupportsDelete:  canDelete,
	}, nil
}

func (a NanoVMSInfrastructureAdapter) Observe(
	ctx context.Context,
	realized RealizedResource,
) (InfrastructureObservation, error) {
	if err := a.validateInvocation(ctx); err != nil {
		return InfrastructureObservation{}, err
	}
	if !nanoVMSNonemptyIdentity(realized.ID) || !nanoVMSNonemptyIdentity(realized.DesiredResourceID) {
		return InfrastructureObservation{}, fmt.Errorf("NanoVMS observation requires complete realized identity")
	}
	if realized.TargetID != a.TargetID {
		return InfrastructureObservation{}, fmt.Errorf(
			"realized target %q does not match NanoVMS target %q",
			realized.TargetID,
			a.TargetID,
		)
	}
	if realized.Provider != a.Provider {
		return InfrastructureObservation{}, fmt.Errorf(
			"realized provider %q does not match NanoVMS provider %q",
			realized.Provider,
			a.Provider,
		)
	}
	if !nanoVMSNonemptyIdentity(realized.ExternalID) {
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
		ConfigDigest:       sandbox.ConfigDigest,
		State:              sandbox.Status,
		Fresh:              true,
	}, nil
}

// validateInvocation rejects local precondition failures before any provider I/O.
func (a NanoVMSInfrastructureAdapter) validateInvocation(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("NanoVMS invocation context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if a.Transport == nil || !nanoVMSNonemptyIdentity(a.TargetID) || !nanoVMSNonemptyIdentity(a.Provider) {
		return fmt.Errorf("NanoVMS adapter requires transport and exact target/provider identity")
	}
	return nil
}

func (a NanoVMSInfrastructureAdapter) unknownMutation(id, lookup string) InfrastructureApplyResult {
	return InfrastructureApplyResult{
		Outcome: InfrastructureApplyUnknown,
		ExternalOperation: &ExternalOperationRef{
			Provider: a.Provider, TargetID: a.TargetID, ExternalID: id, LookupKind: lookup,
		},
	}
}

func (a NanoVMSInfrastructureAdapter) Apply(
	ctx context.Context,
	action PlannedResourceAction,
	desired *DesiredResource,
	realized *RealizedResource,
) (InfrastructureApplyResult, error) {
	if err := a.validateInvocation(ctx); err != nil {
		return InfrastructureApplyResult{}, err
	}
	switch action.Action {
	case ReconcileCreate:
		if desired == nil || realized != nil {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS CREATE requires desired state only")
		}
		if !nanoVMSNonemptyIdentity(string(action.OperationID)) {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS CREATE requires runtime operation identity")
		}
		if desired.Target != a.TargetID {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS CREATE target mismatch")
		}
		if !nanoVMSNonemptyIdentity(desired.ID) || action.DesiredResourceID != desired.ID || action.RealizedResourceID != "" {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS CREATE action and desired identity disagree")
		}
		switch desired.Lifecycle {
		case LifecycleCreateObserve, LifecycleManage, LifecycleOrphanOnRemove, LifecycleDestroyOnExplicitIntent:
		default:
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS CREATE is not authorized by lifecycle %q", desired.Lifecycle)
		}
		if desired.Artifact == nil || *desired.Artifact == "" || a.Artifacts == nil {
			return InfrastructureApplyResult{}, fmt.Errorf(
				"NanoVMS CREATE requires an authorized immutable artifact",
			)
		}
		artifact, found, err := a.Artifacts.ResolveBuildArtifact(ctx, *desired.Artifact)
		if err != nil {
			return InfrastructureApplyResult{}, fmt.Errorf("resolve build artifact: %w", err)
		}
		if !found || artifact.ID != *desired.Artifact || artifact.ImmutableRef == "" {
			return InfrastructureApplyResult{}, fmt.Errorf(
				"NanoVMS CREATE artifact is missing or has no immutable reference",
			)
		}
		sandbox, err := a.Transport.Deploy(ctx, NanoVMSDeployRequest{
			Name:         desired.ID,
			Image:        artifact.ImmutableRef,
			OperationID:  string(action.OperationID),
			ResourceID:   desired.ID,
			ConfigDigest: desired.ConfigDigest,
		})
		if err != nil {
			var unknown *NanoVMSOutcomeUnknownError
			if errors.As(err, &unknown) {
				return InfrastructureApplyResult{
					Outcome: InfrastructureApplyUnknown,
					ExternalOperation: &ExternalOperationRef{
						Provider:   a.Provider,
						TargetID:   a.TargetID,
						ExternalID: string(action.OperationID),
						LookupKind: "byteport-operation",
					},
				}, nil
			}
			return InfrastructureApplyResult{}, err
		}
		if !nanoVMSNonemptyIdentity(sandbox.ID) {
			return a.unknownMutation(string(action.OperationID), "byteport-operation"), nil
		}
		out := RealizedResource{
			ID:                "nanovms:" + sandbox.ID,
			DesiredResourceID: desired.ID,
			TargetID:          a.TargetID,
			Provider:          a.Provider,
			ExternalID:        sandbox.ID,
		}
		return InfrastructureApplyResult{Outcome: InfrastructureApplyRealized, Realized: &out}, nil

	case ReconcileDelete:
		if realized == nil || !nanoVMSNonemptyIdentity(realized.ID) ||
			!nanoVMSNonemptyIdentity(realized.DesiredResourceID) ||
			!nanoVMSNonemptyIdentity(string(action.OperationID)) ||
			action.RealizedResourceID != realized.ID || action.DesiredResourceID != realized.DesiredResourceID ||
			realized.TargetID != a.TargetID || realized.Provider != a.Provider ||
			!nanoVMSNonemptyIdentity(realized.ExternalID) {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS DELETE requires exact action/operation/desired/realized/provider/target/sandbox identities")
		}
		if desired != nil && (desired.ID != realized.DesiredResourceID || desired.Target != a.TargetID || desired.Lifecycle != LifecycleDestroyOnExplicitIntent) {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS DELETE conflicts with the supplied desired resource")
		}
		deleter, supported := a.Transport.(NanoVMSDeletionTransport)
		if !supported {
			return InfrastructureApplyResult{}, fmt.Errorf("NanoVMS DELETE unsupported: Stop does not remove a resource")
		}
		if err := deleter.Delete(ctx, realized.ExternalID, string(action.OperationID)); err != nil {
			var unknown *NanoVMSOutcomeUnknownError
			if errors.As(err, &unknown) {
				return a.unknownMutation(realized.ExternalID, "nanovms-sandbox"), nil
			}
			return InfrastructureApplyResult{}, err
		}
		_, found, err := a.Transport.Observe(ctx, realized.ExternalID)
		if err != nil || found {
			return a.unknownMutation(realized.ExternalID, "nanovms-sandbox"), nil
		}
		return InfrastructureApplyResult{
			Outcome: InfrastructureApplyRealized,
			ExternalOperation: &ExternalOperationRef{
				Provider: a.Provider, TargetID: a.TargetID,
				ExternalID: realized.ExternalID, LookupKind: "nanovms-delete",
			},
			Observation: &InfrastructureObservation{
				RealizedResourceID: realized.ID, TargetID: a.TargetID, Provider: a.Provider,
				State: "absent", Fresh: true,
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
