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

	managed, err := managedOwnerSet(in.ManagedOwners)
	if err != nil {
		return Plan{}, err
	}

	liveResources, err := flatten(in.Live)
	if err != nil {
		return Plan{}, err
	}
	desiredResources, err := flatten(in.Desired)
	if err != nil {
		return Plan{}, err
	}

	plan := Plan{Conflicts: identityConflicts(in.Live, in.Desired)}
	for _, key := range sortedResourceKeys(liveResources, desiredResources) {
		live, hasLive := liveResources[key]
		desired, hasDesired := desiredResources[key]
		appendResourceDiff(&plan, live, hasLive, desired, hasDesired, managed)
	}
	sortPlan(&plan)

	return plan, nil
}

func managedOwnerSet(owners []string) (map[string]struct{}, error) {
	managed := make(map[string]struct{}, len(owners))
	for _, owner := range owners {
		owner = strings.TrimSpace(owner)
		if owner == "" {
			return nil, errors.New("managed owner must not be empty")
		}
		managed[owner] = struct{}{}
	}
	return managed, nil
}

func identityConflicts(live, desired Snapshot) []Conflict {
	var conflicts []Conflict
	if live.Provider != desired.Provider {
		conflicts = append(conflicts, Conflict{
			ResourceType: "provider",
			Key:          live.Provider,
			Reason:       "provider identity differs",
		})
	}
	if live.AccountID != desired.AccountID {
		conflicts = append(conflicts, Conflict{
			ResourceType: "account",
			Key:          live.AccountID,
			Reason:       "account identity differs",
		})
	}
	if live.Tunnel.ID != desired.Tunnel.ID {
		conflicts = append(conflicts, Conflict{
			ResourceType: "tunnel",
			Key:          live.Tunnel.ID,
			Reason:       "tunnel identity differs; import/adoption cannot replace a tunnel",
		})
	}
	return conflicts
}

func sortedResourceKeys(live, desired map[string]resource) []string {
	keys := make(map[string]struct{}, len(live)+len(desired))
	for key := range live {
		keys[key] = struct{}{}
	}
	for key := range desired {
		keys[key] = struct{}{}
	}

	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	return ordered
}

func appendResourceDiff(
	plan *Plan,
	live resource,
	hasLive bool,
	desired resource,
	hasDesired bool,
	managed map[string]struct{},
) {
	switch {
	case !hasLive:
		appendAddition(plan, desired, managed)
	case !hasDesired:
		appendRemoval(plan, live, managed)
	default:
		appendExistingDiff(plan, live, desired, managed)
	}
}

func appendAddition(plan *Plan, desired resource, managed map[string]struct{}) {
	if !isManaged(managed, desired.owner) {
		plan.Conflicts = append(plan.Conflicts, Conflict{
			ResourceType: desired.resourceType,
			Key:          desired.key,
			Reason:       "desired state adds a resource owned outside the managed scope",
			DesiredOwner: desired.owner,
		})
		return
	}

	plan.Changes = append(plan.Changes, Change{
		Kind: ChangeAdd, ResourceType: desired.resourceType, Key: desired.key,
		Owner: desired.owner, After: desired.fingerprint,
	})
}

func appendRemoval(plan *Plan, live resource, managed map[string]struct{}) {
	if !isManaged(managed, live.owner) {
		plan.Conflicts = append(plan.Conflicts, conflictFor(
			live,
			"",
			"desired state would remove a resource owned outside the managed scope",
		))
		return
	}

	plan.Changes = append(plan.Changes, Change{
		Kind: ChangeRemove, ResourceType: live.resourceType, Key: live.key,
		Owner: live.owner, Before: live.fingerprint,
	})
}

func appendExistingDiff(
	plan *Plan,
	live resource,
	desired resource,
	managed map[string]struct{},
) {
	if live.owner != desired.owner {
		plan.Conflicts = append(plan.Conflicts, Conflict{
			ResourceType: live.resourceType,
			Key:          live.key,
			Reason:       "ownership transfer requires explicit migration; planner will not infer it",
			LiveOwner:    live.owner,
			DesiredOwner: desired.owner,
		})
		return
	}
	if live.fingerprint == desired.fingerprint {
		return
	}
	if !isManaged(managed, live.owner) {
		plan.Conflicts = append(plan.Conflicts, conflictFor(
			live,
			desired.owner,
			"desired state would mutate a resource owned outside the managed scope",
		))
		return
	}

	plan.Changes = append(plan.Changes, Change{
		Kind: ChangeUpdate, ResourceType: live.resourceType, Key: live.key,
		Owner: live.owner, Before: live.fingerprint, After: desired.fingerprint,
	})
}

func sortPlan(plan *Plan) {
	sort.Slice(plan.Changes, func(i, j int) bool {
		if plan.Changes[i].ResourceType == plan.Changes[j].ResourceType {
			if plan.Changes[i].Key == plan.Changes[j].Key {
				return plan.Changes[i].Kind < plan.Changes[j].Kind
			}
			return plan.Changes[i].Key < plan.Changes[j].Key
		}
		return plan.Changes[i].ResourceType < plan.Changes[j].ResourceType
	})
	sort.Slice(plan.Conflicts, func(i, j int) bool {
		if plan.Conflicts[i].ResourceType == plan.Conflicts[j].ResourceType {
			if plan.Conflicts[i].Key == plan.Conflicts[j].Key {
				return plan.Conflicts[i].Reason < plan.Conflicts[j].Reason
			}
			return plan.Conflicts[i].Key < plan.Conflicts[j].Key
		}
		return plan.Conflicts[i].ResourceType < plan.Conflicts[j].ResourceType
	})
}

func ValidateSnapshot(snapshot Snapshot) error {
	if err := validateSnapshotIdentity(snapshot); err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(snapshot.Provider), "cloudflare") {
		if err := validateCloudflareSnapshot(snapshot); err != nil {
			return err
		}
	}
	if err := validateAccessPolicies(snapshot.AccessApps); err != nil {
		return err
	}

	_, err := flatten(snapshot)
	return err
}

func validateSnapshotIdentity(snapshot Snapshot) error {
	if strings.TrimSpace(snapshot.Provider) == "" {
		return errors.New("provider is required")
	}
	if strings.TrimSpace(snapshot.AccountID) == "" {
		return errors.New("account_id is required")
	}
	if strings.TrimSpace(snapshot.Tunnel.ID) == "" {
		return errors.New("tunnel.id is required")
	}
	return nil
}

func validateCloudflareSnapshot(snapshot Snapshot) error {
	if err := validateCloudflareRoutes(snapshot.Routes); err != nil {
		return err
	}
	return validateCloudflareDNS(snapshot.DNS)
}

func validateCloudflareDNS(records []DNSRecord) error {
	for _, record := range records {
		if strings.TrimSpace(record.ID) == "" {
			return fmt.Errorf(
				"cloudflare DNS record %q %q requires provider record id",
				record.Name,
				record.Type,
			)
		}
	}
	return nil
}

func validateAccessPolicies(apps []AccessApp) error {
	for _, app := range apps {
		if err := validateAppPolicyPrecedence(app); err != nil {
			return err
		}
	}
	return nil
}

func validateAppPolicyPrecedence(app AccessApp) error {
	seen := make(map[int]string, len(app.PolicyRefs))
	for _, policy := range app.PolicyRefs {
		if policy.Precedence <= 0 {
			return fmt.Errorf(
				"access app %q policy %q requires positive precedence",
				app.Domain,
				policy.ID,
			)
		}
		if previous, exists := seen[policy.Precedence]; exists {
			return fmt.Errorf(
				"access app %q policies %q and %q share precedence %d",
				app.Domain,
				previous,
				policy.ID,
				policy.Precedence,
			)
		}
		seen[policy.Precedence] = policy.ID
	}
	return nil
}

func validateCloudflareRoutes(routes []Route) error {
	if len(routes) == 0 {
		return errors.New("cloudflare snapshot requires at least one ingress route")
	}

	catchAll, err := cloudflareCatchAllPosition(routes)
	if err != nil {
		return err
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

func cloudflareCatchAllPosition(routes []Route) (int, error) {
	catchAll := -1
	for position, route := range routes {
		if strings.TrimSpace(route.Service) == "" {
			return -1, fmt.Errorf("cloudflare route at position %d has empty service", position)
		}
		if !isCatchAllRoute(route) {
			continue
		}
		if catchAll >= 0 {
			return -1, errors.New("cloudflare snapshot contains multiple catch-all routes")
		}
		catchAll = position
	}
	if catchAll < 0 {
		return -1, errors.New("cloudflare snapshot requires a catch-all route")
	}
	return catchAll, nil
}

func isCatchAllRoute(route Route) bool {
	return strings.TrimSpace(route.Hostname) == "" && strings.TrimSpace(route.Path) == ""
}

type resource struct {
	resourceType string
	key          string
	owner        string
	fingerprint  string
}

func flatten(snapshot Snapshot) (map[string]resource, error) {
	out := make(map[string]resource)

	for position, route := range snapshot.Routes {
		item, err := routeResource(position, route)
		if err != nil {
			return nil, err
		}
		if err := addResource(out, item); err != nil {
			return nil, err
		}
	}
	for _, record := range snapshot.DNS {
		item, err := dnsResource(record)
		if err != nil {
			return nil, err
		}
		if err := addResource(out, item); err != nil {
			return nil, err
		}
	}
	for _, app := range snapshot.AccessApps {
		item, err := accessAppResource(app)
		if err != nil {
			return nil, err
		}
		if err := addResource(out, item); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func addResource(out map[string]resource, item resource) error {
	if strings.TrimSpace(item.owner) == "" {
		return fmt.Errorf("%s %q has no owner", item.resourceType, item.key)
	}
	compound := item.resourceType + ":" + item.key
	if _, exists := out[compound]; exists {
		return fmt.Errorf("duplicate %s key %q", item.resourceType, item.key)
	}
	out[compound] = item
	return nil
}

func routeResource(position int, route Route) (resource, error) {
	fingerprint, err := stableJSON(struct {
		Position    int    `json:"position"`
		Service     string `json:"service"`
		Audience    string `json:"audience,omitempty"`
		AccessAppID string `json:"access_app_id,omitempty"`
	}{position, route.Service, route.Audience, route.AccessAppID})
	if err != nil {
		return resource{}, err
	}
	return resource{"route", routeKey(route), route.Owner, fingerprint}, nil
}

func dnsResource(record DNSRecord) (resource, error) {
	fingerprint, err := stableJSON(struct {
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
		return resource{}, err
	}
	return resource{"dns", dnsKey(record), record.Owner, fingerprint}, nil
}

func accessAppResource(app AccessApp) (resource, error) {
	policies := append([]PolicyRef(nil), app.PolicyRefs...)
	fingerprint, err := stableJSON(struct {
		ID         string      `json:"id"`
		Type       string      `json:"type"`
		Audience   string      `json:"audience,omitempty"`
		PolicyRefs []PolicyRef `json:"policy_refs,omitempty"`
		SecretRef  *SecretRef  `json:"machine_secret,omitempty"`
	}{app.ID, app.Type, app.Audience, policies, app.MachineSecret})
	if err != nil {
		return resource{}, err
	}
	return resource{
		"access_app",
		strings.ToLower(strings.TrimSpace(app.Domain)),
		app.Owner,
		fingerprint,
	}, nil
}

func routeKey(route Route) string {
	host := strings.ToLower(strings.TrimSpace(route.Hostname))
	if host == "" {
		host = "<catch-all>"
	}
	return host + "|" + strings.TrimSpace(route.Path)
}

func dnsKey(record DNSRecord) string {
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
	encoded, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func isManaged(managed map[string]struct{}, owner string) bool {
	_, ok := managed[owner]
	return ok
}

func conflictFor(item resource, desiredOwner, reason string) Conflict {
	return Conflict{
		ResourceType: item.resourceType,
		Key:          item.key,
		Reason:       reason,
		LiveOwner:    item.owner,
		DesiredOwner: desiredOwner,
	}
}
