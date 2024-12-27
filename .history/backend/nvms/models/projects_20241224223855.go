package models

type Project struct {
    UUID        string       `gorm:"type:text;primaryKey"`
    Owner       string       `gorm:"type:text;not null;index"` 
    User        User         `gorm:"foreignKey:Owner;references:UUID"`
    Name        string       `gorm:"type:text;not null"`
    Id          string       `gorm:"type:text;uniqueIndex;not null"`
    
    RepositoryID string      `gorm:"type:text;index"`
    Repository   Repository  `gorm:"foreignKey:RepositoryID;references:ID"`
    
    NvmsConfig  NVMS        `gorm:"serializer:json"`
    Readme      string      `gorm:"type:text"`
    Description string      `gorm:"type:text"`
    LastUpdated time.Time   `gorm:"autoUpdateTime"`
    Platform    string      `gorm:"type:text"`
    AccessURL   string      `gorm:"type:text"`
    Type        string      `gorm:"type:text"`
    
    DeploymentsJSON string              `gorm:"type:jsonb;column:deployments"`
    Deployments     map[string]Instance `gorm:"-"`
}

func (p *Project) BeforeSave(tx *gorm.DB) error {
    if p.Deployments != nil {
        data, err := json.Marshal(p.Deployments)
        if err != nil {
            return err
        }
        p.DeploymentsJSON = string(data)
    }
    return nil
}

func (p *Project) AfterFind(tx *gorm.DB) error {
    if p.DeploymentsJSON != "" {
        return json.Unmarshal([]byte(p.DeploymentsJSON), &p.Deployments)
    }
    p.Deployments = make(map[string]Instance)
    return nil
}