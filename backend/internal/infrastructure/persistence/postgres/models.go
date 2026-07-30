package postgres

import (
	"time"

	"gorm.io/gorm"
)

// DeploymentModel represents the GORM model for deployments
type DeploymentModel struct {
	// UUID is the canonical deployment identity shared with the public model and
	// production schema. Keep it as the GORM primary key so inserts target the
	// existing uuid-only table instead of asking PostgreSQL to return a legacy
	// integer id column.
	UUID         string  `gorm:"type:uuid;primaryKey"`
	Name         string  `gorm:"not null"`
	Owner        string  `gorm:"index;not null"`
	ProjectUUID  *string `gorm:"index"`
	Status       string  `gorm:"index;not null"`
	Providers    string  `gorm:"type:jsonb"` // JSON-encoded map
	Services     string  `gorm:"type:jsonb"` // JSON-encoded array
	EnvVars      string  `gorm:"type:jsonb"` // JSON-encoded map
	BuildConfig  string  `gorm:"type:jsonb"` // JSON-encoded struct
	CostInfo     string  `gorm:"type:jsonb"` // JSON-encoded struct
	Metadata     string  `gorm:"type:jsonb"` // JSON-encoded deployment metadata
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeployedAt   *time.Time
	TerminatedAt *time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for GORM
func (DeploymentModel) TableName() string {
	return "deployments"
}
