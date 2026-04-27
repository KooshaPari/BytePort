package models

// add owning user uuid
type Instance struct {
	UUID           string   `gorm:"type:text;primaryKey"`
	Name           string   `gorm:"not null"`
	Status         string   `gorm:"not null"`
  
	RootProjectUUID string  `gorm:"not null"` // Foreign key for Project.UUID
	RootProject    Project  `gorm:"foreignKey:RootProjectUUID;references:UUID"` // Establishes the relationshipR
	 
	Resources	  []AWSResource `gorm:"foreignKey:ResUUID;references:UUID"`

}