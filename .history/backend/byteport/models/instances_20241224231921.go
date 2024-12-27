package models

// add owning user uuid
type Instance struct {
	UUID           string   `gorm:"type:text;primaryKey"`
	Name           string   `gorm:"not null"`
	Status         string   `gorm:"not null"`
 
	User           User     `gorm:"foreignKey:Owner;references:UUID"`
	 
	RootProject    Project  `gorm:"foreignKey:RootProjectUUID;references:UUID"` // Establishes the relationshipR
	 
	Resources	  []AWSResource `gorm:"foreignKey:ResUUID;references:UUID"`

}