package models

// add owning user uuid

type Instance struct {
	UUID           string   `gorm:"type:text;primaryKey"`
	Name           string   `gorm:"not null"`
	Status         string   `gorm:"not null"`
	Owner          string   
	 
	Resources	  []AWSResource 

}