package models

type Project struct {
	UUID        string     `gorm:"type:text;primaryKey"`
	
	User        User       `gorm:"foreignKey:Owner;references:UUID"`
	Name        string     `gorm:"not null"`
	Id          string     `gorm:"not null"` // Fixed missing backtick
	Repository  Repository `gorm:"foreignKey:id;references:id"`
	Description string
	LastUpdated string
	Status      string
	Type        string
	Instances   []Instance `gorm:"foreignKey:RootProjectUUID;references:UUID"` // Correctly link Instances to Project
}