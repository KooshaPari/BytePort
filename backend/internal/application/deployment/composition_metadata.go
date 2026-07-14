package deployment

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func validateCompositionMetadata(digest, artifactRef string) error {
	if digest != "" {
		const prefix = "sha256:"
		encoded := strings.TrimPrefix(digest, prefix)
		if encoded == digest || len(encoded) != 64 {
			return fmt.Errorf("composition_digest must be sha256 followed by 64 hexadecimal characters")
		}
		if _, err := hex.DecodeString(encoded); err != nil {
			return fmt.Errorf("composition_digest must be sha256 followed by 64 hexadecimal characters")
		}
	}
	if len(artifactRef) > 512 || strings.IndexFunc(artifactRef, func(r rune) bool { return r < 0x20 || r == 0x7f }) >= 0 {
		return fmt.Errorf("artifact_ref must be at most 512 characters and contain no control characters")
	}
	return nil
}
