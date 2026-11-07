package routes

import (
    "errors"
    "testing"
)

// Test-only executable specification for BP-AD-01.
//
// This is not production implementation and does not count as product
// acceptance. It checks that the proposed operation identity/recovery semantics
// are internally coherent before developer agents implement them.

type recoveryOpState string

const (
    recoveryPlanned     recoveryOpState = "PLANNED"
    recoveryApplying    recoveryOpState = "APPLYING"
    recoveryUnknown     recoveryOpState = "UNKNOWN"
    recoveryReconciling recoveryOpState = "RECONCILING"
    recoveryRealized    recoveryOpState = "REALIZED"
)

type recoveryOp struct {
    ID               string
    Fingerprint      string
    State            recoveryOpState
    ProviderResource string
}

type recoveryJournal struct {
    ops map[string]recoveryOp
}

func newRecoveryJournal() *recoveryJournal {
    return &recoveryJournal{ops: map[string]recoveryOp{}}
}

func (j *recoveryJournal) begin(id, fingerprint string) (recoveryOp, error) {
    if existing, ok := j.ops[id]; ok {
        if existing.Fingerprint != fingerprint {
            return recoveryOp{}, errors.New("operation id reused with different request fingerprint")
        }
        return existing, nil
    }
    op := recoveryOp{ID: id, Fingerprint: fingerprint, State: recoveryPlanned}
    j.ops[id] = op
    return op, nil
}

func (j *recoveryJournal) setState(id string, state recoveryOpState) {
    op := j.ops[id]
    op.State = state
    j.ops[id] = op
}

func (j *recoveryJournal) recordProviderResource(id, resource string) {
    op := j.ops[id]
    op.ProviderResource = resource
    j.ops[id] = op
}

func (j *recoveryJournal) recoverAfterRestart() {
    for id, op := range j.ops {
        if op.State == recoveryApplying {
            op.State = recoveryUnknown
            j.ops[id] = op
        }
    }
}

func TestRecoveryOperationSameIDSameFingerprintReusesDurableIntent(t *testing.T) {
    j := newRecoveryJournal()
    first, err := j.begin("op-1", "fp-A")
    if err != nil {
        t.Fatal(err)
    }
    second, err := j.begin("op-1", "fp-A")
    if err != nil {
        t.Fatal(err)
    }
    if first != second || len(j.ops) != 1 {
        t.Fatalf("same operation retry created divergent intent: first=%+v second=%+v count=%d", first, second, len(j.ops))
    }
}

func TestRecoveryOperationSameIDDifferentFingerprintRejects(t *testing.T) {
    j := newRecoveryJournal()
    if _, err := j.begin("op-1", "fp-A"); err != nil {
        t.Fatal(err)
    }
    if _, err := j.begin("op-1", "fp-B"); err == nil {
        t.Fatal("same OperationID with different request fingerprint must reject")
    }
}

func TestRecoveryApplyingBecomesUnknownAfterRestart(t *testing.T) {
    j := newRecoveryJournal()
    if _, err := j.begin("op-1", "fp-A"); err != nil {
        t.Fatal(err)
    }
    j.setState("op-1", recoveryApplying)

    // A restart cannot truthfully turn an in-flight remote mutation into
    // success or absence without observing/reconciling provider reality.
    j.recoverAfterRestart()

    if got := j.ops["op-1"].State; got != recoveryUnknown {
        t.Fatalf("post-restart applying state = %s, want UNKNOWN", got)
    }
}

func TestRecoveryProviderIdentitySurvivesUnknownAndReconcile(t *testing.T) {
    j := newRecoveryJournal()
    if _, err := j.begin("op-1", "fp-A"); err != nil {
        t.Fatal(err)
    }
    j.setState("op-1", recoveryApplying)
    j.recordProviderResource("op-1", "provider-777")
    j.recoverAfterRestart()

    op := j.ops["op-1"]
    if op.State != recoveryUnknown || op.ProviderResource != "provider-777" {
        t.Fatalf("recovery lost identity: %+v", op)
    }

    j.setState("op-1", recoveryReconciling)
    j.setState("op-1", recoveryRealized)
    if got := j.ops["op-1"]; got.State != recoveryRealized || got.ProviderResource != "provider-777" {
        t.Fatalf("reconcile did not preserve exact provider identity: %+v", got)
    }
}
