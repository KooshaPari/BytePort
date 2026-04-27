package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)
type DeploymentMap map[string]Instance

func (dm *DeploymentMap) Scan(value interface{}) error {
    if value == nil {
        *dm = make(DeploymentMap)
        return nil
    }

    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("failed to unmarshal DeploymentMap value: %v", value)
    }

    map_ob := make(DeploymentMap)
    err := json.Unmarshal(bytes, &map_ob)
    if err != nil {
        return err
    }

    *dm = map_ob
    return nil
}

func (dm DeploymentMap) Value() (driver.Value, error) {
    if dm == nil {
        return nil, nil
    }
    return json.Marshal(dm)
}

type Project struct {
	UUID  string `gorm:"type:text;primaryKey"`
	Owner string `gorm:"type:text;not null;index"`
	User  User   `gorm:"foreignKey:Owner;references:UUID"`
	Name  string `gorm:"type:text;not null"`
	Id    string `gorm:"type:text;uniqueIndex;not null"`

	RepositoryID string     `gorm:"type:text;index"`
	Repository   Repository `gorm:"foreignKey:RepositoryID;references:ID"`

	NvmsConfig  NVMS      `gorm:"serializer:json"`
	Readme      string    `gorm:"type:text"`
	Description string    `gorm:"type:text"`
	LastUpdated time.Time `gorm:"autoUpdateTime"`
	Platform    string    `gorm:"type:text"`
	AccessURL   string    `gorm:"type:text"`
	Type        string    `gorm:"type:text"`

	DeploymentsJSON string              `gorm:"type:jsonb;column:deployments"`
	Deployments     DeploymentMap 
}

func (p *Project) BeforeSave(tx *gorm.DB) error {
	if p.Deployments != nil {
		data, err := json.Marshal(p.Deployments)
		if err != nil {
			return err
		}
		p.DeploymentsJSON = string(data)
	}
	return nil
}

func (p *Project) AfterFind(tx *gorm.DB) error {
	if p.DeploymentsJSON != "" {
		return json.Unmarshal([]byte(p.DeploymentsJSON), &p.Deployments)
	}
	p.Deployments = make(map[string]Instance)
	return nil
}