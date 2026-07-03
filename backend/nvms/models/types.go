package models

// NVMS, Service and BuildPack were previously aliased to an external
// github.com/kooshapari/nanovms/parser package as part of a planned
// cross-repo consolidation (epic B10). That upstream package was never
// published (the nanovms repo has no parser package), which broke
// `go build`/`go vet` for this module. Restored as local types matching
// the pre-extraction definitions until the upstream extraction lands.
type BuildPack struct {
	Name            string            `yaml:"NAME"`                       // Name of the buildpack
	DetectFiles     []string          `yaml:"DETECT_FILES,omitempty"`     // Files that indicate this buildpack should be used
	Packages        []string          `yaml:"PACKAGES"`                   // System packages needed
	PreBuild        []string          `yaml:"PRE_BUILD"`                  // Commands to run before building
	Build           []string          `yaml:"BUILD"`                      // Build commands
	Start           string            `yaml:"START"`                      // Command to start the application
	RuntimeVersions map[string]string `yaml:"RUNTIME_VERSIONS,omitempty"` // Maps language version files to install commands
	EnvVars         map[string]string `yaml:"ENV_VARS"`                   // Required environment variables
}

type NVMS struct {
	Name        string    `yaml:"NAME"`
	Description string    `yaml:"DESCRIPTION"`
	Services    []Service `yaml:"SERVICES"`
}

type Service struct {
	Name      string            `yaml:"NAME"`
	Path      string            `yaml:"PATH"`
	Port      int               `yaml:"PORT"`
	Build     []string          `yaml:"BUILD,omitempty"`     // Keep for custom build overrides
	Env       map[string]string `yaml:"ENV,omitempty"`       // Additional environment variables
	BuildPack *BuildPack        `yaml:"BUILDPACK,omitempty"` // Optional, will use auto-detection if not specified
	Runtime   string            `yaml:"RUNTIME,omitempty"`   // Optional version override
}

type AWSConfig struct {
	Region   string
	Services []AWSServiceConfig
}
type AWSServiceConfig struct {
	Type       string
	Engine     string
	Mode       string
	Replicas   int
	Size       string
	Name       string
	Partitions int
}

type AWSResource struct {
	ID         string                   `json:"id"`
	Type       string                   `json:"type"` // e.g., "ec2", "alb", "targetgroup"
	Name       string                   `json:"name"`
	ARN        string                   `json:"arn"`
	Status     string                   `json:"status"`
	Region     string                   `json:"region"`
	Tags       map[string]string        `json:"tags"`
	Properties map[string]interface{}   `json:"properties"`
	Associates []AWSResourceAssociation `json:"associates"`
	Service    string                   `json:"service"`
}

type AWSResourceAssociation struct {
	ResourceID string `json:"resource_id"`
	Type       string `json:"type"` // e.g., "attachment", "dependency"
	Role       string `json:"role"` // e.g., "target", "source"
}
