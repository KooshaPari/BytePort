package models

import (
	"time"
)

type Project struct {
	UUID        string     `json:"uuid"`
	Owner       string     `json:"owner"`
	User        User       `json:"user"`
	Name        string     `json:"name"`

	RepositoryID string     `json:"repository_id"`
	Repository   Repository `json:"repository"`

	NvmsConfig  NVMS       `json:"nvms_config"`
	Readme      string     `json:"readme"`
	Description string     `json:"description"`
	LastUpdated time.Time  `json:"last_updated"`
	Platform    string     `json:"platform"`
	AccessURL   string     `json:"access_url"`
	Type        string     `json:"type"`

	deployments map[string]Instance `json:"deployments,omitempty"`
}
 func (p *Project) GetDeploy() map[string]Instance{}
 func (p *Project) SetDeploy(deploy map[string]Instance){
	
 }
 func (p *Project) AppendDeploy(key string, deploy Instance){
	p.deployments[key] = deploy
 }