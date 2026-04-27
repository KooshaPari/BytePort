package models

import (
	"gorm.io/gorm"
)

type BuildPack struct {
    ID              uint             `gorm:"primaryKey;autoIncrement"`
    Name            string           `yaml:"NAME"`
    DetectFiles     []string         `yaml:"DETECT_FILES,omitempty" gorm:"type:text"` // store as a JSON/text field
    Packages        []string         `yaml:"PACKAGES" gorm:"type:text"`
    PreBuild        []string         `yaml:"PRE_BUILD" gorm:"type:text"`
    Build           []string         `yaml:"BUILD" gorm:"type:text"`
    Start           string           `yaml:"START"`
    RuntimeVersions map[string]string `yaml:"RUNTIME_VERSIONS,omitempty" gorm:"-"`
    EnvVars         map[string]string `yaml:"ENV_VARS" gorm:"-"`
}

type NVMS struct {
    gorm.Model
    Name        string    `yaml:"NAME"`
    Description string    `yaml:"DESCRIPTION"`
    Services    []Service `yaml:"SERVICES" gorm:"-"`
}

type Service struct {
    ID        uint             `gorm:"primaryKey;autoIncrement"`
    Name      string           `yaml:"NAME"`
    Path      string           `yaml:"PATH"`
    Port      int              `yaml:"PORT"`
    Build     []string         `yaml:"BUILD,omitempty" gorm:"type:text"`
    Env       map[string]string `yaml:"ENV,omitempty" gorm:"-"`
    BuildPack *BuildPack       `yaml:"BUILDPACK,omitempty" gorm:"-"`
    Runtime   string           `yaml:"RUNTIME,omitempty"`
}

type AWSConfig struct {
    gorm.Model
    Region   string
    Services []AWSServiceConfig `gorm:"-"`
}

type AWSServiceConfig struct {
    ID        uint   `gorm:"primaryKey;autoIncrement"`
    Type      string
    Engine    string
    Mode      string
    Replicas  int
    Size      string
    Name      string
    Partitions int
}

type AWSResource struct {
    gorm.Model
    ID       string `json:"id" gorm:"<-:create;primaryKey"`
    Type     string `json:"type"`
    Name     string `json:"name"`
    ARN      string `json:"arn"`
    Status   string `json:"status"`
    Region   string `json:"region"`
    Service  string `json:"service"`
}

type AWSResourceAssociation struct {
    gorm.Model
    ResourceID string `json:"resource_id"`
    Type       string `json:"type"`
    Role       string `json:"role"`
}
