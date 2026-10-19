package sharedingress

import "time"

// Snapshot is a secret-free observation or desired-state projection of shared
// ingress. It intentionally stores provider identifiers and secret references,
// never credential values.
type Snapshot struct {
	ObservedAt time.Time   `json:"observed_at"`
	Source     string      `json:"source"`
	Provider   string      `json:"provider"`
	AccountID  string      `json:"account_id"`
	Tunnel     Tunnel      `json:"tunnel"`
	Routes     []Route     `json:"routes"`
	DNS        []DNSRecord `json:"dns"`
	AccessApps []AccessApp `json:"access_apps"`
	Connectors []Connector `json:"connectors"`
}

type Tunnel struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status,omitempty"`
	ConfigVersion int    `json:"config_version,omitempty"`
}

type SecretRef struct {
	Store   string `json:"store"`
	Key     string `json:"key"`
	Version string `json:"version,omitempty"`
}

type Route struct {
	Hostname    string `json:"hostname,omitempty"`
	Path        string `json:"path,omitempty"`
	Service     string `json:"service"`
	Owner       string `json:"owner"`
	Audience    string `json:"audience,omitempty"`
	AccessAppID string `json:"access_app_id,omitempty"`
}

type DNSRecord struct {
	// ID is the provider's stable record identity. Cloudflare snapshots require
	// it because several records may legitimately share the same name and type.
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	ZoneID  string `json:"zone_id,omitempty"`
	Owner   string `json:"owner"`
}

type AccessApp struct {
	ID            string      `json:"id"`
	Domain        string      `json:"domain"`
	Type          string      `json:"type"`
	Owner         string      `json:"owner"`
	Audience      string      `json:"audience,omitempty"`
	PolicyRefs    []PolicyRef `json:"policy_refs,omitempty"`
	MachineSecret *SecretRef  `json:"machine_secret,omitempty"`
}

type PolicyRef struct {
	ID           string   `json:"id"`
	Name         string   `json:"name,omitempty"`
	Decision     string   `json:"decision"`
	Precedence   int      `json:"precedence"`
	IncludeKinds []string `json:"include_kinds,omitempty"`
	RequireKinds []string `json:"require_kinds,omitempty"`
	ExcludeKinds []string `json:"exclude_kinds,omitempty"`
}

type Connector struct {
	ID            string   `json:"id"`
	Arch          string   `json:"arch,omitempty"`
	Version       string   `json:"version,omitempty"`
	ConfigVersion int      `json:"config_version,omitempty"`
	Colos         []string `json:"colos,omitempty"`
	Status        string   `json:"status,omitempty"`
}
