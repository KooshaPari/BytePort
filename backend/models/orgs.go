package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Org represents a multi-tenant organization.
type Org struct {
	gorm.Model
	UUID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"uuid,omitempty"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	OwnerID     string    `gorm:"type:uuid;not null;index" json:"owner_id"`
	Owner       User      `gorm:"foreignKey:OwnerID;references:UUID" json:"owner,omitempty"`
	Members     []OrgMember `json:"members,omitempty" gorm:"foreignKey:OrgID;references:UUID"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// OrgMember represents a user's membership in an organization with a role.
type OrgMember struct {
	gorm.Model
	UUID   string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"uuid,omitempty"`
	OrgID  string `gorm:"type:uuid;not null;index:idx_org_user,unique" json:"org_id"`
	UserID string `gorm:"type:uuid;not null;index:idx_org_user,unique" json:"user_id"`
	Role   string `gorm:"type:varchar(50);not null;default:'member'" json:"role"` // owner, admin, member

	Org  Org  `gorm:"foreignKey:OrgID;references:UUID" json:"org,omitempty"`
	User User `gorm:"foreignKey:UserID;references:UUID" json:"user,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// OrgRole constants for RBAC.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// ResourceType enum for fine-grained permissions.
type ResourceType string

const (
	ResourceDeployment ResourceType = "deployment"
	ResourceProject    ResourceType = "project"
	ResourceSecret     ResourceType = "secret"
	ResourceOrg        ResourceType = "org"
	ResourceProvider   ResourceType = "provider"
)

// Action enum for permission checks.
type Action string

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// rolePermissions maps roles to resource/action pairs they can perform.
var rolePermissions = map[string]map[ResourceType][]Action{
	RoleOwner: {
		ResourceDeployment: {ActionCreate, ActionRead, ActionUpdate, ActionDelete},
		ResourceProject:    {ActionCreate, ActionRead, ActionUpdate, ActionDelete},
		ResourceSecret:     {ActionCreate, ActionRead, ActionUpdate, ActionDelete},
		ResourceOrg:        {ActionCreate, ActionRead, ActionUpdate, ActionDelete},
		ResourceProvider:   {ActionCreate, ActionRead, ActionUpdate, ActionDelete},
	},
	RoleAdmin: {
		ResourceDeployment: {ActionCreate, ActionRead, ActionUpdate, ActionDelete},
		ResourceProject:    {ActionCreate, ActionRead, ActionUpdate, ActionDelete},
		ResourceSecret:     {ActionRead, ActionUpdate},
		ResourceOrg:        {ActionRead, ActionUpdate},
		ResourceProvider:   {ActionCreate, ActionRead, ActionUpdate},
	},
	RoleMember: {
		ResourceDeployment: {ActionCreate, ActionRead, ActionUpdate},
		ResourceProject:    {ActionCreate, ActionRead, ActionUpdate},
		ResourceSecret:     {ActionRead},
		ResourceOrg:        {ActionRead},
		ResourceProvider:   {ActionRead},
	},
}

// Can returns true if the role permits the given action on the given resource type.
func Can(role string, resource ResourceType, action Action) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	for _, a := range perms[resource] {
		if a == action {
			return true
		}
	}
	return false
}

// HasOrgRole returns true if a user has the given role (or higher) in an org.
func HasOrgRole(role string, required string) bool {
	rank := map[string]int{RoleOwner: 3, RoleAdmin: 2, RoleMember: 1}
	return rank[role] >= rank[required]
}

// BeforeCreate hooks for automatic UUID generation.
func (o *Org) BeforeCreate(tx *gorm.DB) error {
	if o.UUID == "" {
		o.UUID = uuid.New().String()
	}
	return nil
}

func (om *OrgMember) BeforeCreate(tx *gorm.DB) error {
	if om.UUID == "" {
		om.UUID = uuid.New().String()
	}
	return nil
}
