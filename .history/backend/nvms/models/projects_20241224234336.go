package models

import (
	"time"
)

type Project struct {
	UUID  string  
	Owner string 
	User  User    
	Name  string  
	Id    string  

	RepositoryID string      
	Repository   Repository  

	NvmsConfig  NVMS      `gorm:"serializer:json"`
	Readme      string    `gorm:"type:text"`
	Description string    `gorm:"type:text"`
	LastUpdated time.Time `gorm:"autoUpdateTime"`
	Platform    string    `gorm:"type:text"`
	AccessURL   string    `gorm:"type:text"`
	Type        string    `gorm:"type:text"`

	DeploymentsJSON string              `gorm:"type:jsonb;column:deployments"`
	Deployments     map[string]Instance 
}
 