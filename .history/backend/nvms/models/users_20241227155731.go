package models

import "time"

type User struct {
	UUID        string       
	Name        string    
	Email       string      
	Password    string       
	AwsCreds    AwsCreds    
	LLMCreds LLM  
	Portfolio   Portfolio    
	Git         Git         
	Projects    []Project  
	Instances   []Instance  
}

type AwsCreds struct {
	AccessKeyID     string `gorm:"column:access_key_id"`
	SecretAccessKey string `gorm:"column:secret_access_key"`
}

type LLM struct {
	Provider   string `gorm:"column:Provider"`
	Providers map[string]AIProvider`gorm:"serializer:json"`

}
type AIProvider struct {
	Modal string `gorm:"column:modal"`
	APIKey string `gorm:"column:api_key"`
}
type Portfolio struct {
	RootEndpoint string `gorm:"column:root_endpoint"`
	APIKey       string `gorm:"column:api_key"`
}

type Git struct {
	Token              string    `gorm:"column:access_token"`
	RefreshToken       string    `gorm:"column:refresh_token"`
	TokenExpiry        time.Time `gorm:"column:token_expiry"`
	RefreshTokenExpiry time.Time `gorm:"column:refresh_token_expiry"` // User-specific GitHub App installation ID
	Repositories       []string  `gorm:"-"`                           // A list of repository names (optional, for frontend display)
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
	AwsCreds    AwsCreds    `gorm:"embedded;embeddedPrefix:aws_"`
	OpenAICreds OpenAICreds `gorm:"embedded;embeddedPrefix:Lll"`
	Portfolio   Portfolio   `gorm:"embedded;embeddedPrefix:portfolio_"`
	Git         Git         `gorm:"embedded;embeddedPrefix:git_"`
}