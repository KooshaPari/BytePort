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

type Project struct {
	UUID string `gorm:"type:text;primaryKey"`
	Name string `gorm:"not null"`
	Description string `gorm:"not null"`
	LastUpdated string `gorm:"not null"`
	Status string `gorm:"not null"`
	Type string `gorm:"not null"`
	Instances []VMInstance `gorm:"foreignKey:Project"`
}