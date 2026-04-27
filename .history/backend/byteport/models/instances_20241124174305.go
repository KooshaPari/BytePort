package models

// config for gorm based on js struct below
/* interface VMInstance {
		Name: String;
		Status: String;
		OS: String;
		RootProject: Project;
		LastUpdated: String;
	}*/
type Instance struct {
	Name string `gorm:"not null"`
	Status string `gorm:"not null"`
	OS string `gorm:"not null"`
	RootProject Project `gorm:"embedded;embeddedPrefix:root_project_"`
	LastUpdated string `gorm:"not null"`
	
}