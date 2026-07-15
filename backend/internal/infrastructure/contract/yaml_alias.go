// Package contract provides schemathesis-style API contract testing for BytePort.
//
// This package wires the OpenAPI spec into a property-based testing harness that
// verifies API behavior against its specification, including status codes, schema
// conformance, and edge cases for boundary values.
//
// yamlAlias re-exports yaml.Node through an exported name so callers can construct
// YAML nodes for use in OpenAPI examples and assertions without importing
// gopkg.in/yaml.v3 directly across package boundaries.
package contract

import "gopkg.in/yaml.v3"

// YAMLNode is a thin alias to yaml.Node used by the contract testing harness.
// All fields and methods of yaml.Node are available through this alias.
type YAMLNode = yaml.Node

// YMarshaler is an alias to yaml.Marshaler to make the contract package self-contained.
type YMarshaler = yaml.Marshaler

// YUnmarshaler is an alias to yaml.Unmarshaler.
type YUnmarshaler = yaml.Unmarshaler

// yamlNode is the lowercase alias used internally by the harness tests.
type yamlNode = yaml.Node
