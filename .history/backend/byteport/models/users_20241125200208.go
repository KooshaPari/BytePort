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
    Provider        string        `gorm:"column:provider"` // e.g., "github", "gitlab", "bitbucket"
    BaseURL         string        `gorm:"column:base_url"` // API base URL for Git provider
    AuthMethod      string        `gorm:"column:auth_method"` // e.g., "token", "ssh"
    AuthKey         string        `gorm:"column:auth_key"` // OAuth token, PAT, or SSH key
    DefaultRepoURL  string        `gorm:"column:default_repo_url"` // Default repository to clone or interact with
    Repositories    []Repository  `gorm:"foreignKey:Owner;references:UUID"` // Associated repositories
    TargetDirectory string        `gorm:"column:target_directory"` // Target directory for operations
}

type Repository struct {
    ID              uint   `gorm:"primaryKey"` // Internal ID for this entry
    Owner           string `gorm:"not null"`  // User or organization that owns the repository
    Name            string `gorm:"not null"`  // Repository name (e.g., "example-repo")
    CloneURL        string `gorm:"not null"`  // Repository clone URL
    Visibility      string `gorm:"not null"`  // e.g., "public", "private"
    LastSyncedAt    string `gorm:"type:timestamp"` // Last time this repository was synced
    GitCredentialID uint   // Link to a specific Git credential set if needed
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