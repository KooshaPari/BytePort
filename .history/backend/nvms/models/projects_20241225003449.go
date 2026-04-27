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
	Platform    string     
	AccessURL   string    
	Type        string    

          
	Deployments     map[string]Instance `json:`
}
 