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
    InstallationID  string `gorm:"column:installation_id"`   // User-specific GitHub App installation ID
    Repositories    []string `gorm:"-"`                     // A list of repository names (optional, for frontend display)
    SelectedRepoURL string `gorm:"column:selected_repo_url"`// The specific repository the user selected to integrate
    Token           string `gorm:"column:token"`            // Temporary installation token (for scoped access)
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