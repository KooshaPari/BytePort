package models

// add owning user uuid
type Instance struct {
	UUID           string   `gorm:"type:text;primaryKey"`
	Name           string   `gorm:"not null"`
	Status         string   `gorm:"not null"`
	ResUUID		string   `gorm:"not null"  `
	ResourcesJSON   string        `gorm:"type:jsonb;column:resources"`
    Resources       []AWSResource `gorm:"foreignKey:InstanceID;references:UUID"`
}

 