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
		// BUILD and ENV are parsed as untrusted manifest data, but projecting a
		// desired graph must not authorize execution or materialize secret values.
		// The graph digest therefore captures only declarative service placement
		// inputs that are safe at this stage.
		declarative := struct {
			Name    string `json:"name"`
			Path    string `json:"path"`
			Port    int    `json:"port"`
			Runtime string `json:"runtime,omitempty"`
		}{
			Name: svc.Name, Path: svc.Path, Port: svc.Port, Runtime: svc.Runtime,
		}
		raw, err := json.Marshal(declarative)
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
