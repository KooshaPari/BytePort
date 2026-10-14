package sharedingress

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ChangeKind string

const (
	ChangeAdd    ChangeKind = "add"
	ChangeUpdate ChangeKind = "update"
	ChangeRemove ChangeKind = "remove"
)

type Change struct {
	Kind         ChangeKind `json:"kind"`
	ResourceType string     `json:"resource_type"`
	Key          string     `json:"key"`
	Owner        string     `json:"owner"`
	Before       string     `json:"before,omitempty"`
	After        string     `json:"after,omitempty"`
}

type Conflict struct {
	ResourceType string `json:"resource_type"`
	Key          string `json:"key"`
	Reason       string `json:"reason"`
	LiveOwner    string `json:"live_owner,omitempty"`
	DesiredOwner string `json:"desired_owner,omitempty"`
}

type Plan struct {
	Changes   []Change   `json:"changes"`
	Conflicts []Conflict `json:"conflicts"`
}

func (p Plan) Safe() bool { return len(p.Conflicts) == 0 }

type PlanInput struct {
	Live          Snapshot
	Desired       Snapshot
	ManagedOwners []string
}

// BuildPlan computes a deterministic, side-effect-free drift plan. It has no
// apply path by design: issue #438 starts with import/plan qualification only.
func BuildPlan(in PlanInput) (Plan, error) {
	if err := ValidateSnapshot(in.Live); err != nil {
		return Plan{}, fmt.Errorf("live snapshot: %w", err)
	}
	if err := ValidateSnapshot(in.Desired); err != nil {
		return Plan{}, fmt.Errorf("desired snapshot: %w", err)
	}

	managed := make(map[string]struct{}, len(in.ManagedOwners))
	for _, owner := range in.ManagedOwners {
		owner = strings.TrimSpace(owner)
		if owner == "" {
			return Plan{}, errors.New("managed owner must not be empty")
		}
		managed[owner] = struct{}{}
	}

	var p Plan
	if in.Live.Provider != in.Desired.Provider {
		p.Conflicts = append(p.Conflicts, Conflict{
			ResourceType: "provider",
			Key:          in.Live.Provider,
			Reason:       "provider identity differs",
		})
	}
	if in.Live.AccountID != in.Desired.AccountID {
		p.Conflicts = append(p.Conflicts, Conflict{
			ResourceType: "account",
			Key:          in.Live.AccountID,
			Reason:       "account identity differs",
		})
	}
	if in.Live.Tunnel.ID != in.Desired.Tunnel.ID {
		p.Conflicts = append(p.Conflicts, Conflict{
			ResourceType: "tunnel",
			Key:          in.Live.Tunnel.ID,
			Reason:       "tunnel identity differs; import/adoption cannot replace a tunnel",
		})
	}

	liveResources, err := flatten(in.Live)
	if err != nil {
		return Plan{}, err
	}
	desiredResources, err := flatten(in.Desired)
	if err != nil {
		return Plan{}, err
	}

	keys := make(map[string]struct{}, len(liveResources)+len(desiredResources))
	for k := range liveResources {
		keys[k] = struct{}{}
	}
	for k := range desiredResources {
		keys[k] = struct{}{}
	}
	ordered := make([]string, 0, len(keys))
	for k := range keys {
		ordered = append(ordered, k)
	}
	sort.Strings(ordered)

	for _, compoundKey := range ordered {
		live, hasLive := liveResources[compoundKey]
		desired, hasDesired := desiredResources[compoundKey]

		switch {
		case hasLive && !hasDesired:
			if !isManaged(managed, live.owner) {
				p.Conflicts = append(p.Conflicts, conflictFor(live, "", "desired state would remove a resource owned outside the managed scope"))
				continue
			}
			p.Changes = append(p.Changes, Change{
				Kind: ChangeRemove, ResourceType: live.resourceType, Key: live.key,
				Owner: live.owner, Before: live.fingerprint,
			})

		case !hasLive && hasDesired:
			if !isManaged(managed, desired.owner) {
				p.Conflicts = append(p.Conflicts, conflictFor(desired, desired.owner, "desired state adds a resource owned outside the managed scope"))
				continue
			}
			p.Changes = append(p.Changes, Change{
				Kind: ChangeAdd, ResourceType: desired.resourceType, Key: desired.key,
				Owner: desired.owner, After: desired.fingerprint,
			})

		case hasLive && hasDesired:
			if live.owner != desired.owner {
				p.Conflicts = append(p.Conflicts, Conflict{
					ResourceType: live.resourceType,
					Key:          live.key,
					Reason:       "ownership transfer requires explicit migration; planner will not infer it",
					LiveOwner:    live.owner,
					DesiredOwner: desired.owner,
				})
				continue
			}
			if live.fingerprint == desired.fingerprint {
				continue
			}
			if !isManaged(managed, live.owner) {
				p.Conflicts = append(p.Conflicts, conflictFor(live, desired.owner, "desired state would mutate a resource owned outside the managed scope"))
				continue
			}
			p.Changes = append(p.Changes, Change{
				Kind: ChangeUpdate, ResourceType: live.resourceType, Key: live.key,
				Owner: live.owner, Before: live.fingerprint, After: desired.fingerprint,
			})
		}
	}

	sort.Slice(p.Changes, func(i, j int) bool {
		if p.Changes[i].ResourceType == p.Changes[j].ResourceType {
			if p.Changes[i].Key == p.Changes[j].Key {
				return p.Changes[i].Kind < p.Changes[j].Kind
			}
			return p.Changes[i].Key < p.Changes[j].Key
		}
		return p.Changes[i].ResourceType < p.Changes[j].ResourceType
	})
	sort.Slice(p.Conflicts, func(i, j int) bool {
		if p.Conflicts[i].ResourceType == p.Conflicts[j].ResourceType {
			if p.Conflicts[i].Key == p.Conflicts[j].Key {
				return p.Conflicts[i].Reason < p.Conflicts[j].Reason
			}
			return p.Conflicts[i].Key < p.Conflicts[j].Key
		}
		return p.Conflicts[i].ResourceType < p.Conflicts[j].ResourceType
	})

	return p, nil
}

func ValidateSnapshot(s Snapshot) error {
	if strings.TrimSpace(s.Provider) == "" {
		return errors.New("provider is required")
	}
	if strings.TrimSpace(s.AccountID) == "" {
		return errors.New("account_id is required")
	}
	if strings.TrimSpace(s.Tunnel.ID) == "" {
		return errors.New("tunnel.id is required")
	}
	_, err := flatten(s)
	return err
}

type resource struct {
	resourceType string
	key          string
	owner        string
	fingerprint  string
}

func flatten(s Snapshot) (map[string]resource, error) {
	out := make(map[string]resource)

	add := func(r resource) error {
		if strings.TrimSpace(r.owner) == "" {
			return fmt.Errorf("%s %q has no owner", r.resourceType, r.key)
		}
		compound := r.resourceType + ":" + r.key
		if _, exists := out[compound]; exists {
			return fmt.Errorf("duplicate %s key %q", r.resourceType, r.key)
		}
		out[compound] = r
		return nil
	}

	for _, r := range s.Routes {
		key := routeKey(r)
		fp, err := stableJSON(struct {
			Service     string `json:"service"`
			Audience    string `json:"audience,omitempty"`
			AccessAppID string `json:"access_app_id,omitempty"`
		}{r.Service, r.Audience, r.AccessAppID})
		if err != nil {
			return nil, err
		}
		if err := add(resource{"route", key, r.Owner, fp}); err != nil {
			return nil, err
		}
	}

	for _, r := range s.DNS {
		key := strings.ToLower(strings.TrimSpace(r.Name)) + "|" + strings.ToUpper(strings.TrimSpace(r.Type))
		fp, err := stableJSON(struct {
			Content string `json:"content"`
			Proxied bool   `json:"proxied"`
			ZoneID  string `json:"zone_id,omitempty"`
		}{r.Content, r.Proxied, r.ZoneID})
		if err != nil {
			return nil, err
		}
		if err := add(resource{"dns", key, r.Owner, fp}); err != nil {
			return nil, err
		}
	}

	for _, a := range s.AccessApps {
		key := strings.ToLower(strings.TrimSpace(a.Domain))
		policies := append([]PolicyRef(nil), a.PolicyRefs...)
		sort.Slice(policies, func(i, j int) bool { return policies[i].ID < policies[j].ID })
		fp, err := stableJSON(struct {
			ID         string      `json:"id"`
			Type       string      `json:"type"`
			Audience   string      `json:"audience,omitempty"`
			PolicyRefs []PolicyRef `json:"policy_refs,omitempty"`
			SecretRef  *SecretRef  `json:"machine_secret,omitempty"`
		}{a.ID, a.Type, a.Audience, policies, a.MachineSecret})
		if err != nil {
			return nil, err
		}
		if err := add(resource{"access_app", key, a.Owner, fp}); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func routeKey(r Route) string {
	host := strings.ToLower(strings.TrimSpace(r.Hostname))
	if host == "" {
		host = "<catch-all>"
	}
	return host + "|" + strings.TrimSpace(r.Path)
}

func stableJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func isManaged(managed map[string]struct{}, owner string) bool {
	_, ok := managed[owner]
	return ok
}

func conflictFor(r resource, desiredOwner, reason string) Conflict {
	return Conflict{
		ResourceType: r.resourceType,
		Key:          r.key,
		Reason:       reason,
		LiveOwner:    r.owner,
		DesiredOwner: desiredOwner,
	}
}
