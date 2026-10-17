package models

import "time"

// RuntimeOperationRecord is the durable product journal for externally visible
// runtime mutation. It is additive to legacy Project storage.
type RuntimeOperationRecord struct {
	ID string `gorm:"primaryKey;size:128" json:"id"`
	Owner string `gorm:"index;size:255" json:"owner"`
	Fingerprint string `gorm:"size:128;not null" json:"fingerprint"`
	State RuntimeOperationState `gorm:"size:32;not null" json:"state"`
	Target string `gorm:"size:255" json:"target"`
	ProjectUUID string `gorm:"index;size:255" json:"project_uuid,omitempty"`
	ProviderResourceID string `gorm:"size:255" json:"provider_resource_id,omitempty"`
	LastError string `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
