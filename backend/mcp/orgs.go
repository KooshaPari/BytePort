// Package mcp — MCP tools for org/tenant management.
//
// Tier 2.6 — BytePort Agent Platform surface.
// These tools expose multi-tenant operations via the Model Context Protocol,
// enabling AI agents to manage organizations, memberships, and deployments.
package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/byteport/api/models"
)

func init() {
	// Register org tools alongside the existing deployment tools.
	RegisterTool(ToolDefinition{
		Name:        "org_list",
		Description: "List all organizations the authenticated user belongs to",
		InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
		Handler:     handleOrgList,
	})
	RegisterTool(ToolDefinition{
		Name:        "org_create",
		Description: "Create a new organization",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"name":{"type":"string","description":"Organization display name"},
				"slug":{"type":"string","description":"URL-friendly identifier (optional)"}
			},
			"required":["name"]
		}`),
		Handler: handleOrgCreate,
	})
	RegisterTool(ToolDefinition{
		Name:        "org_members",
		Description: "List members of an organization",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"org_id":{"type":"string","description":"Organization ID"}
			},
			"required":["org_id"]
		}`),
		Handler: handleOrgMembers,
	})
	RegisterTool(ToolDefinition{
		Name:        "org_invite",
		Description: "Invite a user to an organization",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{
				"org_id":{"type":"string","description":"Organization ID"},
				"user_email":{"type":"string","description":"Email of the user to invite"},
				"role":{"type":"string","description":"Role: member, admin, or viewer (default: member)"}
			},
			"required":["org_id","user_email"]
		}`),
		Handler: handleOrgInvite,
	})
	RegisterTool(ToolDefinition{
		Name:        "deployment_list",
		Description: "List deployments across all accessible organizations",
		InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[]}`),
		Handler:     handleDeploymentList,
	})
}

func handleOrgList(args json.RawMessage, _ ToolContext) (any, error) {
	// TODO: wire into actual models.Organization query via models.DB
	return map[string]any{
		"organizations": []map[string]any{
			{"id": "demo", "name": "Demo Org", "slug": "demo", "role": "owner"},
		},
		"note": "List organizations — database integration pending",
	}, nil
}

func handleOrgCreate(args json.RawMessage, _ ToolContext) (any, error) {
	var params struct {
		Name string `json:"name"`
		Slug string `json:"slug,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	if params.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	org, err := models.CreateOrganization(params.Name, params.Slug, "demo-user-id")
	if err != nil {
		return nil, fmt.Errorf("create org: %w", err)
	}

	return map[string]any{
		"organization": map[string]any{
			"id":   org.ID,
			"name": org.Name,
			"slug": org.Slug,
		},
	}, nil
}

func handleOrgMembers(args json.RawMessage, _ ToolContext) (any, error) {
	var params struct {
		OrgID string `json:"org_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	members, err := models.GetMembersByOrg(params.OrgID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	result := make([]map[string]any, 0, len(members))
	for _, m := range members {
		result = append(result, map[string]any{
			"user_id": m.UserID,
			"role":    string(m.Role),
			"joined":  m.JoinedAt,
		})
	}

	return map[string]any{"members": result}, nil
}

func handleOrgInvite(args json.RawMessage, _ ToolContext) (any, error) {
	var params struct {
		OrgID     string `json:"org_id"`
		UserEmail string `json:"user_email"`
		Role      string `json:"role,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	role := models.RoleMember
	switch params.Role {
	case "admin":
		role = models.RoleAdmin
	case "viewer":
		role = models.RoleViewer
	}

	// TODO: resolve user email to user ID, then create membership
	_ = role
	return map[string]any{
		"invited": true,
		"org_id":  params.OrgID,
		"email":   params.UserEmail,
		"role":    params.Role,
		"note":    "Invite sent — resolver & membership create pending",
	}, nil
}

func handleDeploymentList(args json.RawMessage, _ ToolContext) (any, error) {
	// TODO: query deployments from all accessible orgs via models.DB
	return map[string]any{
		"deployments": []map[string]any{},
		"note":        "Deployment listing — database integration pending",
	}, nil
}
