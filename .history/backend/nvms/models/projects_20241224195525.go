package models

type Project struct {
	UUID        string     `gorm:"type:text;primaryKey"`
	Owner       string     `gorm:"not null"` // References User.UUID
	User        User       `gorm:"foreignKey:Owner;references:UUID"`
	Name        string     `gorm:"not null"`
	Id          string     `gorm:"not null"` // Fixed missing backtick
	Repository  Repository `gorm:"foreignKey:id;references:id"`
	NvmsConfig  NVMS
	Readme      string
	Description string
	LastUpdated string
	Platform    string
	AccessURL   string
	Type        string
	Deployments   []Instance `gorm:"foreignKey:RootProjectUUID;references:UUID"` // Correctly link Instances to Project
}