package models

// add owning user uuid

type Instance struct {
	UUID           string   `gorm:"type:text;primaryKey"`
	Name           string    
	Status         string   
	Owner          string   
	 
	Resources	  []AWSResource 

}