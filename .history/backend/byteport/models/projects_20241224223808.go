package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Project struct {
	 UUID        string       `gorm:"type:text;primaryKey"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
    
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

	Resources   ResourceMap  `gorm:"type:jsonb"`
    Deployments []Deployment `gorm:"foreignKey:ProjectID;references:UUID"`
}

type Deployment struct {
	 ID          string       `gorm:"type:text;primaryKey"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
    

	Name      string `gorm:"type:text;not null"`
	ProjectID string `gorm:"type:text;index;not null"`

	InstanceID string   `gorm:"type:text;index"`
	Instance   Instance `gorm:"foreignKey:InstanceID;references:UUID"`
 	Resources   ResourceMap  `gorm:"type:jsonb"`
	Status  string `gorm:"type:text;not null;default:'pending'"`
	Version string `gorm:"type:text"`
}
type ResourceMap map[string]interface{}

func (rm *ResourceMap) Scan(value interface{}) error {
    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("failed to unmarshal ResourceMap value")
    }

    return json.Unmarshal(bytes, &rm)
}

func (rm ResourceMap) Value() (driver.Value, error) {
    if rm == nil {
        return nil, nil
    }
    return json.Marshal(rm)
}
