package models

import (
	"github.com/kooshapari/nanovms/parser"
)

// Type aliases for NVMS parser types from phenotype-tooling/nanovms/parser.
// These were extracted as part of cross-repo consolidation (epic B10).
type (
	NVMS      = parser.NVMS
	Service   = parser.Service
	BuildPack = parser.BuildPack
)

// AWS-specific types remain local as they are BytePort-specific.
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
