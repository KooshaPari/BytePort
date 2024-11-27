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
	
}