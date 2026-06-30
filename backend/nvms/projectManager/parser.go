package projectManager

import (
	"github.com/kooshapari/nanovms/parser"
)

// parseNVMSConfig delegates to the shared parser from phenotype-tooling/nanovms/parser.
// This package was extracted as part of cross-repo consolidation (epic B10).
func parseNVMSConfig(yamlContent string) (*parser.NVMS, error) {
	return parser.ParseNVMSConfig(yamlContent)
}
