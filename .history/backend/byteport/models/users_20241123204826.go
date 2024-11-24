package models

import (
	"gorm.io/gorm"
)

type User struct {
    gorm.Model
	UUID     string `gorm:"unique;not null"`
    Name     string `gorm:"not null"`
    Email    string `gorm:"unique;not null"`
    Password string `gorm:"not null"`

	SystemCreds  SystemCreds  `gorm:"embedded;embeddedPrefix:system_"`
    AwsCreds     AwsCreds     `gorm:"embedded;embeddedPrefix:aws_"`
    OpenAICreds  OpenAICreds  `gorm:"embedded;embeddedPrefix:openai_"`
    Portfolio    Portfolio    `gorm:"embedded;embeddedPrefix:portfolio_"`
    Git          Git          `gorm:"embedded;embeddedPrefix:git_"`
}

type AwsCreds struct {
    AccessKeyID     string `gorm:"column:access_key_id"`
    SecretAccessKey string `gorm:"column:secret_access_key"`
}

type OpenAICreds struct {
    APIKey string `gorm:"column:api_key"`
}

type Portfolio struct {
    RootEndpoint string `gorm:"column:root_endpoint"`
    APIKey       string `gorm:"column:api_key"`
}

type Git struct {
    RepoURL         string `gorm:"column:repo_url"`
    AuthMethod      string `gorm:"column:auth_method"`
    AuthKey         string `gorm:"column:auth_key"`
    TargetDirectory string `gorm:"column:target_directory"`
}
type SystemCreds struct {
	EncryptedToken string `gorm:"column:token"`
	SecretKey      string `gorm:"column:secret_key"`
}