package models

type Project struct {
	UUID        string              `gorm:"type:text;primaryKey"`
	Owner       string              `gorm:"type:text;not null"` // References User.UUID
	User        User                `gorm:"foreignKey:Owner;references:UUID"`
	Name        string              `gorm:"type:text;not null"`
	Id          string              `gorm:"type:text;not null"`
	Repository  Repository          `gorm:"foreignKey:id;references:id"`
	NvmsConfig  NVMS                `gorm:"type:jsonb"`
	Readme      string              `gorm:"type:text"`
	Description string              `gorm:"type:text"`
	LastUpdated string              `gorm:"type:text"`
	Platform    string              `gorm:"type:text"`
	AccessURL   string              `gorm:"type:text"`
	Type        string              `gorm:"type:text"`
	Deployments struct{} `gorm:"type:jsonb"`
}
