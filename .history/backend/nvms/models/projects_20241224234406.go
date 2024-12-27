package models

import (
	"time"
)

type Project struct {
	UUID  string  
	Owner string 
	User  User    
	Name  string  
	 

	RepositoryID string      
	Repository   Repository  

	NvmsConfig  NVMS      
	Readme      string     
	Description string    
	LastUpdated time.Time  
	Platform    string    `gorm:"type:text"`
	AccessURL   string    `gorm:"type:text"`
	Type        string    `gorm:"type:text"`

	DeploymentsJSON string              `gorm:"type:jsonb;column:deployments"`
	Deployments     map[string]Instance 
}
 