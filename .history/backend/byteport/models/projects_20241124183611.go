package models

type Project struct {
	UUID        string      `gorm:"type:text;primaryKey"`
	Owner       string      `gorm:"not null"` // References User.UUID
	User        User        `gorm:"foreignKey:Owner;references:UUID"`
	Name        string      `gorm:"not null"`
	Description string
	LastUpdated string
	Status      string
	Type        string
	Instances   []Instance  `gorm:"foreignKey:RootProjectUUID;references:UUID"` // Correctly link Instances to Project
}