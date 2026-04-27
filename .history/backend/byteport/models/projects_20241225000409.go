package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	UUID  string `gorm:"type:text;primaryKey"`
	Owner string `gorm:"type:text;not null;index"`
	User  User   `gorm:"-"`
	Name  string `gorm:"type:text;not null"`
	 

	RepositoryID string     `gorm:"type:text;index"`
	Repository   Repository `gorm:"foreignKey:RepositoryID;references:ID"`

	        
	
	Readme      string    `gorm:"type:text"`
	Description string    `gorm:"type:text"`
	LastUpdated time.Time `gorm:"autoUpdateTime"`
	Platform    string    `gorm:"type:text"`
	AccessURL   string    `gorm:"type:text"`
	Type        string    `gorm:"type:text"`

	DeploymentsJSON string              `gorm:"type:jsonb;column:deployments"`
	deployments     map[string]Instance  
}
func(p *Project) getDeploy() map[string]Instance{
	return p.deployments
}
func(p *Project) setDeploy(deploy map[string]Instance){
func (p *Project) BeforeSave(tx *gorm.DB) error {
	if p.deployments != nil {
		data, err := json.Marshal(p.deployments)
		if err != nil {
			return err
		}
		p.DeploymentsJSON = string(data)
	}
	if p.UUID == "" {
		fmt.Println("Generating UUID")
		p.UUID = uuid.New().String()
	}
	return nil
}

func (p *Project) AfterFind(tx *gorm.DB) error {
	if p.DeploymentsJSON != "" {
		return json.Unmarshal([]byte(p.DeploymentsJSON), &p.deployments)
	}
	p.deployments = make(map[string]Instance)
	return nil
}