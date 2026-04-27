package models

// add owning user uuid
type Instance struct {
	UUID         string   `gorm:"type:text;primaryKey"`
	Name         string   `gorm:"not null"`
	Status       string   `gorm:"not null"`
	OS           string   `gorm:"not null"`
	Owner        string   `gorm:"not null"` // This references User.UUID
	User         User     `gorm:"foreignKey:Owner;references:UUID"` // Explicit foreign key relationship
	RootProject  Project  `gorm:"foreignKey:RootProjectUUID;references:UUID"` // Correctly link to Project
	LastUpdated  string   `gorm:"not null"`
}