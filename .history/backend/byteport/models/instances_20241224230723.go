package models

// add owning user uuid
type Instance struct {
	UUID           string   `gorm:"type:text;primaryKey"`
	Name           string   `gorm:"not null"`
	Status         string   `gorm:"not null"`
	Owner          string   `gorm:"not null"` // References User.UUID
	User           User     `gorm:"foreignKey:Owner;references:UUID"`
	RootProjectUUID string  `gorm:"not null"` // Foreign key for Project.UUID
	RootProject    Project  `gorm:"foreignKey:RootProjectUUID;references:UUID"` // Establishes the relationshipR
	RUUID string `gorm:"not null"`
	Resources	  []AWSResource `gorm:"foreignKey:InstanceUUID;references:RUUID"`

}