package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// RecoveryManifestToDesiredGraph translates the recovered historical service
// manifest into the mature generalized graph. Target selection is explicit;
// the compatibility layer never invents provider placement.
func RecoveryManifestToDesiredGraph(
	rev ManifestRevision,
	manifest RecoveryManifest,
	targetID string,
) (DesiredResourceGraph, error) {
	if targetID == "" {
		return DesiredResourceGraph{}, fmt.Errorf("target ID is required for historical manifest import")
	}

	resources := make([]DesiredResource, 0, len(manifest.Services))
	for _, svc := range manifest.Services {
		raw, err := json.Marshal(svc)
		if err != nil {
			return DesiredResourceGraph{}, fmt.Errorf("digest service %q: %w", svc.Name, err)
		}
		sum := sha256.Sum256(raw)
		resources = append(resources, DesiredResource{
			ID:           svc.Name,
			Kind:         DesiredResourceService,
			ConfigDigest: hex.EncodeToString(sum[:]),
			Target:       targetID,
			Lifecycle:    LifecycleManage,
		})
	}

	return DesiredResourceGraph{
		ID:        string(rev.ID) + ":graph",
		Manifest:  rev.ID,
		Resources: resources,
	}, nil
}
