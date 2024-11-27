package models

type User struct {
	UUID       string      `gorm:"type:text;primaryKey"`
	Name       string      `gorm:"not null"`
	Email      string      `gorm:"unique;not null"`
	Password   string      `gorm:"not null"`
	AwsCreds   AwsCreds    `gorm:"embedded;embeddedPrefix:aws_"`
	OpenAICreds OpenAICreds `gorm:"embedded;embeddedPrefix:openai_"`
	Portfolio  Portfolio   `gorm:"embedded;embeddedPrefix:portfolio_"`
	Git        Git         `gorm:"embedded;embeddedPrefix:git_"`
	Projects   []Project   `gorm:"foreignKey:Owner;references:UUID"`
	Instances  []Instance  `gorm:"foreignKey:Owner;references:UUID"`
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
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}
type SignupRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}
// contains everything not in signup request but in the original user object
type LinkRequest struct {
    AwsCreds     AwsCreds     `gorm:"embedded;embeddedPrefix:aws_"`
    OpenAICreds  OpenAICreds  `gorm:"embedded;embeddedPrefix:openai_"`
    Portfolio    Portfolio    `gorm:"embedded;embeddedPrefix:portfolio_"`
    Git          Git          `gorm:"embedded;embeddedPrefix:git_"`
}