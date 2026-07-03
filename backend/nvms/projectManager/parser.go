package projectManager

import (
	"fmt"
	"strings"

	"nvms/models"

	"gopkg.in/yaml.v2"
)

// parseNVMSConfig parses an NVMS YAML manifest into models.NVMS.
//
// This previously delegated to an external github.com/kooshapari/nanovms/parser
// package as part of a planned cross-repo consolidation (epic B10), but that
// upstream package was never published, breaking go build/vet. Restored to
// local parsing using the module's own models types until the extraction lands.
func parseNVMSConfig(yamlContent string) (*models.NVMS, error) {
	if strings.TrimSpace(yamlContent) == "" {
		return nil, fmt.Errorf("empty YAML content")
	}

	config := &models.NVMS{}
	if err := yaml.Unmarshal([]byte(yamlContent), config); err != nil {
		return nil, fmt.Errorf("YAML parsing error: %w", err)
	}

	if config.Name == "" {
		return nil, fmt.Errorf("missing required field: NAME")
	}
	if len(config.Services) == 0 {
		return nil, fmt.Errorf("no services defined in YAML")
	}

	found := false
	for _, svc := range config.Services {
		if svc.Path == "" {
			return nil, fmt.Errorf("service %s missing PATH", svc.Name)
		}
		if svc.Port == 0 {
			return nil, fmt.Errorf("service %s missing PORT", svc.Name)
		}
		if svc.Name == "main" {
			if found {
				return nil, fmt.Errorf("service main already defined")
			}
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("service main not defined")
	}

	return config, nil
}
