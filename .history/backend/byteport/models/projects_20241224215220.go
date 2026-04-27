package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	UUID      string `gorm:"type:text;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Owner string `gorm:"type:text;not null;index"`
	User  User   `gorm:"foreignKey:Owner;references:UUID"`

	Name string `gorm:"type:text;not null"`
	Id   string `gorm:"type:text;uniqueIndex;not null"`

	RepositoryID string     `gorm:"type:text;index"`
	Repository   Repository `gorm:"foreignKey:RepositoryID;references:ID"`

	NvmsConfig  NVMS   `gorm:"type:jsonb"`
	Readme      string `gorm:"type:text"`
	Description string `gorm:"type:text"`
	LastUpdated time.Time
	Platform    string `gorm:"type:text"`
	AccessURL   string `gorm:"type:text"`
	Type        string `gorm:"type:text"`

	Deployments []Deployment `gorm:"foreignKey:ProjectID;references:UUID"`
}

type Deployment struct {
	ID        string `gorm:"type:text;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name      string `gorm:"type:text;not null"`
	ProjectID string `gorm:"type:text;index;not null"`

	InstanceID string   `gorm:"type:text;index"`
	Instance   Instance `gorm:"foreignKey:InstanceID;references:UUID"`

	Status  string `gorm:"type:text;not null;default:'pending'"`
	Version string `gorm:"type:text"`
}