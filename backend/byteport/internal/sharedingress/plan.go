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
				p.Conflicts = append(p.Conflicts, conflictFor(
					live,
					"",
					"desired state would remove a resource owned outside the managed scope",
				))
				continue
			}
			p.Changes = append(p.Changes, Change{
				Kind: ChangeRemove, ResourceType: live.resourceType, Key: live.key,
				Owner: live.owner, Before: live.fingerprint,
			})

		case !hasLive && hasDesired:
			if !isManaged(managed, desired.owner) {
				p.Conflicts = append(p.Conflicts, Conflict{
					ResourceType: desired.resourceType,
					Key:          desired.key,
					Reason:       "desired state adds a resource owned outside the managed scope",
					DesiredOwner: desired.owner,
				})
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
				p.Conflicts = append(p.Conflicts, conflictFor(
					live,
					desired.owner,
					"desired state would mutate a resource owned outside the managed scope",
				))
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

	if strings.EqualFold(strings.TrimSpace(s.Provider), "cloudflare") {
		if err := validateCloudflareRoutes(s.Routes); err != nil {
			return err
		}
		for _, record := range s.DNS {
			if strings.TrimSpace(record.ID) == "" {
				return fmt.Errorf(
					"cloudflare DNS record %q %q requires provider record id",
					record.Name,
					record.Type,
				)
			}
		}
	}

	for _, app := range s.AccessApps {
		seenPrecedence := make(map[int]string, len(app.PolicyRefs))
		for _, policy := range app.PolicyRefs {
			if policy.Precedence <= 0 {
				return fmt.Errorf(
					"access app %q policy %q requires positive precedence",
					app.Domain,
					policy.ID,
				)
			}
			if previous, exists := seenPrecedence[policy.Precedence]; exists {
				return fmt.Errorf(
					"access app %q policies %q and %q share precedence %d",
					app.Domain,
					previous,
					policy.ID,
					policy.Precedence,
				)
			}
			seenPrecedence[policy.Precedence] = policy.ID
		}
	}

	_, err := flatten(s)
	return err
}

func validateCloudflareRoutes(routes []Route) error {
	if len(routes) == 0 {
		return errors.New("cloudflare snapshot requires at least one ingress route")
	}

	catchAll := -1
	for position, route := range routes {
		if strings.TrimSpace(route.Service) == "" {
			return fmt.Errorf("cloudflare route at position %d has empty service", position)
		}
		if strings.TrimSpace(route.Hostname) == "" && strings.TrimSpace(route.Path) == "" {
			if catchAll >= 0 {
				return errors.New("cloudflare snapshot contains multiple catch-all routes")
			}
			catchAll = position
		}
	}
	if catchAll < 0 {
		return errors.New("cloudflare snapshot requires a catch-all route")
	}
	if catchAll != len(routes)-1 {
		return fmt.Errorf(
			"cloudflare catch-all route must be last: position %d of %d",
			catchAll,
			len(routes),
		)
	}
	return nil
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

	for position, route := range s.Routes {
		key := routeKey(route)
		fp, err := stableJSON(struct {
			Position    int    `json:"position"`
			Service     string `json:"service"`
			Audience    string `json:"audience,omitempty"`
			AccessAppID string `json:"access_app_id,omitempty"`
		}{position, route.Service, route.Audience, route.AccessAppID})
		if err != nil {
			return nil, err
		}
		if err := add(resource{"route", key, route.Owner, fp}); err != nil {
			return nil, err
		}
	}

	for _, record := range s.DNS {
		key := dnsKey(s.Provider, record)
		fp, err := stableJSON(struct {
			Name    string `json:"name"`
			Type    string `json:"type"`
			Content string `json:"content"`
			Proxied bool   `json:"proxied"`
			ZoneID  string `json:"zone_id,omitempty"`
		}{
			strings.ToLower(strings.TrimSpace(record.Name)),
			strings.ToUpper(strings.TrimSpace(record.Type)),
			record.Content,
			record.Proxied,
			record.ZoneID,
		})
		if err != nil {
			return nil, err
		}
		if err := add(resource{"dns", key, record.Owner, fp}); err != nil {
			return nil, err
		}
	}

	for _, app := range s.AccessApps {
		key := strings.ToLower(strings.TrimSpace(app.Domain))
		// PolicyRefs deliberately retain provider order and explicit precedence.
		// Reordering or changing precedence is authorization drift.
		policies := append([]PolicyRef(nil), app.PolicyRefs...)
		fp, err := stableJSON(struct {
			ID         string      `json:"id"`
			Type       string      `json:"type"`
			Audience   string      `json:"audience,omitempty"`
			PolicyRefs []PolicyRef `json:"policy_refs,omitempty"`
			SecretRef  *SecretRef  `json:"machine_secret,omitempty"`
		}{app.ID, app.Type, app.Audience, policies, app.MachineSecret})
		if err != nil {
			return nil, err
		}
		if err := add(resource{"access_app", key, app.Owner, fp}); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func routeKey(route Route) string {
	host := strings.ToLower(strings.TrimSpace(route.Hostname))
	if host == "" {
		host = "<catch-all>"
	}
	return host + "|" + strings.TrimSpace(route.Path)
}

func dnsKey(provider string, record DNSRecord) string {
	if id := strings.TrimSpace(record.ID); id != "" {
		return "id:" + id
	}
	return strings.Join([]string{
		"fallback",
		strings.TrimSpace(record.ZoneID),
		strings.ToLower(strings.TrimSpace(record.Name)),
		strings.ToUpper(strings.TrimSpace(record.Type)),
		record.Content,
	}, "|")
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
