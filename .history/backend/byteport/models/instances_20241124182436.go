package models

// add owning user uuid
type Instance struct {
	UUID string `gorm:"type:text;primaryKey"`
	Name string `gorm:"not null"`
	Status string `gorm:"not null"`
	OS string `gorm:"not null"`
	User        User    `gorm:"foreignKey:Owner;references:UUID"` // Establish proper relationship
	RootProject Project `gorm:"foreignKey:UUID;references:RootProjectUUID"`
	LastUpdated string  `gorm:"not null"`
	
}