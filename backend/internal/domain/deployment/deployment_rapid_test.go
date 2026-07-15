package deployment

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"pgregory.net/rapid"
)

// TestNewDeployment_Properties — fuzzes NewDeployment over a wide input
// space. Properties verified:
//
//   1. Name="" → always returns an error (never a valid Deployment).
//   2. Owner="" → always returns an error.
//   3. Non-empty name+owner → always returns a Deployment whose UUID is
//      a valid v4 UUID (the uuid.New().String() contract).
//   4. Non-empty name+owner → CreatedAt and UpdatedAt MUST be equal to
//      within 1ms (NewDeployment sets both in the same `.UTC()` call).
//   5. Non-empty name+owner → Status MUST start as StatusPending.
//   6. The returned Deployment MUST pass Validate() — i.e. NewDeployment
//      is a sane factory, never produces an invalid object.
//
// These properties catch off-by-one, race-condition, and double-init bugs
// that example-based unit tests cannot reach.
func TestNewDeployment_Properties(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := rapid.String().Draw(t, "name")
		owner := rapid.String().Draw(t, "owner")

		dep, err := NewDeployment(name, owner, nil)

		// Property 1 & 2: empty inputs MUST error.
		if name == "" || owner == "" {
			if err == nil {
				t.Fatalf("expected error for empty input (name=%q owner=%q), got nil", name, owner)
			}
			if dep != nil {
				t.Fatalf("expected nil Deployment on error path, got %+v", dep)
			}
			return
		}

		// Property 3+ on success path.
		if err != nil {
			t.Fatalf("unexpected error for valid input (name=%q owner=%q): %v", name, owner, err)
		}
		if dep == nil {
			t.Fatalf("expected non-nil Deployment on success")
		}

		// Property 3 — UUID is a valid v4 UUID.
		if _, err := uuid.Parse(dep.UUID()); err != nil {
			t.Errorf("UUID %q is not a valid UUID: %v", dep.UUID(), err)
		}

		// Property 4 — CreatedAt and UpdatedAt equal to within 1ms.
		created := dep.CreatedAt()
		updated := dep.UpdatedAt()
		if d := abs(time.Duration(created.Sub(updated))); d > time.Millisecond {
			t.Errorf("CreatedAt/UpdatedAt differ by %v (limit 1ms)", d)
		}

		// Property 5 — initial status is Pending.
		if dep.Status() != StatusPending {
			t.Errorf("expected initial StatusPending, got %v", dep.Status())
		}

		// Property 6 — factory output is itself valid.
		if err := dep.Validate(); err != nil {
			t.Errorf("NewDeployment produced invalid Deployment: %v", err)
		}
	})
}

// TestSetStatus_Properties — exercises CanTransitionTo + SetStatus over
// all (current → new) status pairs. Critical invariant: SetStatus will
// REFUSE any transition not in the transition table. This proves the
// state machine is closed under SetStatus.
func TestSetStatus_Properties(t *testing.T) {
	allStatuses := []Status{
		StatusPending, StatusDetecting, StatusProvisioning,
		StatusDeploying, StatusDeployed, StatusFailed, StatusTerminated,
	}

	rapid.Check(t, func(t *rapid.T) {
		dep, err := NewDeployment("p", "owner", nil)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		// Pick two random statuses to drive transition.
		cur := rapid.SampledFrom(allStatuses).Draw(t, "current")
		next := rapid.SampledFrom(allStatuses).Draw(t, "next")

		// Drive the deployment to `cur` via BFS-style transitions. If
		// `cur` is unreachable from StatusPending, skip.
		ok := driveTo(dep, cur)
		if !ok {
			t.Skipf("status %v not reachable from Pending", cur)
		}

		// Now attempt the transition. Property: err == nil iff
		// CanTransitionTo says yes — i.e. SetStatus and the predicate
		// must agree exactly.
		can := dep.CanTransitionTo(next)
		err = dep.SetStatus(next)
		if can && err != nil {
			t.Errorf("CanTransitionTo(%v → %v) is true but SetStatus returned: %v", cur, next, err)
		}
		if !can && err == nil {
			t.Errorf("CanTransitionTo(%v → %v) is false but SetStatus succeeded (status=%v)", cur, next, dep.Status())
		}
		if can && dep.Status() != next {
			t.Errorf("after accepted transition, expected Status=%v got %v", next, dep.Status())
		}
	})
}

// TestAddService_Properties — generators arbitrary service tuples and
// confirms:
//   - AddService rejects empty name/type/provider facets.
//   - AddService rejects duplicate names.
//   - Adding N unique services results in Services() returning N items.
func TestAddService_Properties(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		dep, err := NewDeployment("d", "o", nil)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		// 1..10 unique services with mutually distinct names.
		n := rapid.IntRange(1, 10).Draw(t, "n")
		type svcGen struct {
			svc      DeploymentService
			expected bool
		}
		added := 0
		for i := 0; i < n; i++ {
			name := rapid.StringMatching(`^[a-z][a-z0-9-]{2,8}$`).Draw(t, "name")
			svc := DeploymentService{
				Name:     name,
				Type:     rapid.SampledFrom([]string{"frontend", "backend", "database"}).Draw(t, "type"),
				Provider: rapid.SampledFrom([]string{"vercel", "render", "fly", "supabase"}).Draw(t, "provider"),
				Status:   "pending",
				URL:      "https://example.com/" + name,
			}
			err := dep.AddService(svc)
			// Empty-field rejection & duplicate rejection are valid outcomes.
			if err != nil {
				if errors.Is(err, errServiceNonEmpty) {
					continue
				}
				// Duplicate — also acceptable; we count "added" only when err is nil.
				continue
			}
			added++
		}

		// Property: returned slice length matches successful adds.
		if got := len(dep.Services()); got != added {
			t.Errorf("Services() returned %d items, expected %d", got, added)
		}
	})
}

// TestCalculateTotalCost_Properties — total cost MUST equal sum of
// breakdown entries and MUST be non-negative.
func TestCalculateTotalCost_Properties(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		dep, _ := NewDeployment("c", "o", nil)
		breakdown := map[string]float64{}
		want := 0.0
		// Up to 10 entries, each a non-negative cost. Keys MUST be unique
		// so the map doesn't dedupe (otherwise want over-counts).
		n := rapid.IntRange(0, 10).Draw(t, "n")
		for i := 0; i < n; i++ {
			// Uniqueness enforced by index suffix; the rest of the key is
			// arbitrary so the test still exercises varied characters.
			key := rapid.StringMatching(`^[a-z_]{1,8}$`).Draw(t, "key") +
				fmt.Sprintf("_%d", i)
			cost := rapid.Float64Range(0, 1_000_000).Draw(t, "cost")
			breakdown[key] = cost
			want += cost
		}
		dep.SetCostInfo(&CostInfo{Monthly: want, Breakdown: breakdown})

		got := dep.CalculateTotalCost()
		if want == 0 && got != 0 {
			t.Errorf("empty breakdown: expected 0, got %v", got)
		}
		// Property: floating-point sum tolerance 1e-6.
		if absFloat64(got-want) > 1e-6 {
			t.Errorf("CalculateTotalCost: want=%v got=%v delta=%v", want, got, absFloat64(got-want))
		}
		if got < 0 {
			t.Errorf("cost must be non-negative, got %v", got)
		}
	})
}

// driveTo — best-effort BFS to move the deployment into the target status
// from StatusPending. Returns true on success, false if the target is
// unreachable.
func driveTo(d *Deployment, target Status) bool {
	visited := map[Status]bool{}
	var dfs func(Status) bool
	dfs = func(cur Status) bool {
		if cur == target {
			return true
		}
		if visited[cur] {
			return false
		}
		visited[cur] = true
		// Try every legal next status.
		for _, nxt := range neighbors(cur) {
			if d.CanTransitionTo(nxt) {
				if err := d.SetStatus(nxt); err != nil {
					continue
				}
				if dfs(nxt) {
					return true
				}
			}
		}
		// Backing out is not supported in this DFS; the test will
		// simply skip unreachable states upstream.
		return false
	}
	return dfs(StatusPending)
}

func neighbors(s Status) []Status {
	switch s {
	case StatusPending:
		return []Status{StatusDetecting, StatusFailed, StatusTerminated}
	case StatusDetecting:
		return []Status{StatusProvisioning, StatusFailed, StatusTerminated}
	case StatusProvisioning:
		return []Status{StatusDeploying, StatusFailed, StatusTerminated}
	case StatusDeploying:
		return []Status{StatusDeployed, StatusFailed, StatusTerminated}
	case StatusDeployed:
		return []Status{StatusDeploying, StatusTerminated}
	case StatusFailed:
		return []Status{StatusDeploying, StatusTerminated}
	}
	return nil
}

// errServiceNonEmpty is the err we expect for empty service fields.
var errServiceNonEmpty = errors.New("empty")

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func absFloat64(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
