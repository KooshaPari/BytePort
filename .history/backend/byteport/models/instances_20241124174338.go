package models

// config for gorm based on js struct below
/* interface Project {
		UUID: String;
		Name: String;
		Description: String;
		LastUpdated: String;
		Status: String;
		Type: String;
		Instances: VMInstance[];
	}*/
type Instance struct {
	UUID string `gorm:"type:text;primaryKey"`
	Name string `gorm:"not null"`
	Status string `gorm:"not null"`
	OS string `gorm:"not null"`
	RootProject Project `gorm:"embedded;embeddedPrefix:root_project_"`
	LastUpdated string `gorm:"not null"`
	
}