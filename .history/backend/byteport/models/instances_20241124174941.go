package models

// add owning user uuid 
type Instance struct {
	UUID string `gorm:"type:text;primaryKey"`
	Owner string 
	Name string `gorm:"not null"`
	Status string `gorm:"not null"`
	OS string `gorm:"not null"`
	RootProject Project `gorm:"embedded;embeddedPrefix:root_project_"`
	LastUpdated string `gorm:"not null"`
	
}