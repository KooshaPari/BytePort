package models

type Project struct {
	UUID string `gorm:"type:text;primaryKey"`
	Owner string 
	Name string `gorm:"not null"`
	Description string `gorm:"not null"`
	LastUpdated string `gorm:"not null"`
	Status string `gorm:"not null"`
	Type string `gorm:"not null"`
	Instances []Instance `gorm:"foreignKey:Project"`
}