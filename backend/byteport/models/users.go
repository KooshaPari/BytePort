package models

import (
	"encoding/json"
	"strings"
	"time"
)

type User struct {
	UUID      string     `json:"uuid" gorm:"type:text;primaryKey"`
	Name      string     `json:"name" gorm:"not null"`
	Email     string     `json:"email" gorm:"unique;not null"`
	Password  string     `json:"password" gorm:"not null"`
	AwsCreds  AwsCreds   `json:"awsCreds" gorm:"embedded;embeddedPrefix:aws_"`
	LLMConfig LLM        `json:"llmConfig" gorm:"embedded;embeddedPrefix:llm_"`
	Portfolio Portfolio  `json:"portfolio" gorm:"embedded;embeddedPrefix:portfolio_"`
	Git       Git        `json:"git" gorm:"embedded;embeddedPrefix:git_"`
	Projects  []Project  `json:"projects" gorm:"foreignKey:Owner;references:UUID"`
	Instances []Instance `json:"instances" gorm:"foreignKey:Owner;references:UUID"`
}

type LLM struct {
	Provider  string                `json:"provider" gorm:"column:Provider"`
	Providers map[string]AIProvider `json:"providers" gorm:"serializer:json"`
}

// ProviderEntry resolves a provider by name and also reports the key it is
// stored under. Lookup is case-insensitive because the canonical spelling is
// lower case ("openai") while clients have sent "openAI"; a case-sensitive
// lookup missed the entry, so the credential was read as an empty string.
func (l LLM) ProviderEntry(name string) (string, AIProvider, bool) {
	if provider, ok := l.Providers[name]; ok {
		return name, provider, true
	}
	for key, provider := range l.Providers {
		if strings.EqualFold(key, name) {
			return key, provider, true
		}
	}
	return "", AIProvider{}, false
}

// ProviderKey returns the configured provider for name, matching
// case-insensitively. See ProviderEntry for the canonical key.
func (l LLM) ProviderKey(name string) (AIProvider, bool) {
	_, provider, ok := l.ProviderEntry(name)
	return provider, ok
}

// AIProvider describes one configured LLM provider.
//
// The credential is serialised as `api_key`. Clients have shipped camelCase
// `apiKey` for that same field, and encoding/json only matches names that
// differ by letter case ("apiKey" does not match "api_key"), so the value was
// silently dropped and the credential validated as empty, which surfaced as
// "Failed to validate OAI credentials" on every setup attempt.
//
// UnmarshalJSON accepts every spelling that has appeared on the wire; Marshal
// always emits the canonical `api_key` (see the struct tags below).
type AIProvider struct {
	Modal  string `json:"modal" gorm:"column:modal"`
	APIKey string `json:"api_key" gorm:"column:api_key"`
}

// UnmarshalJSON decodes an AIProvider from any of the credential spellings in
// use: `api_key` (canonical), `apiKey`, and the case variants of `apikey`.
func (p *AIProvider) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = AIProvider{
		Modal:  firstStringField(raw, "modal", "Modal"),
		APIKey: firstStringField(raw, "api_key", "apiKey", "apikey", "APIKey", "ApiKey"),
	}
	return nil
}

// firstStringField returns the value of the first key present in raw that
// decodes as a JSON string.
func firstStringField(raw map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		var decoded string
		if err := json.Unmarshal(value, &decoded); err == nil {
			return decoded
		}
	}
	return ""
}

type AwsCreds struct {
	AccessKeyID     string `gorm:"column:access_key_id"`
	SecretAccessKey string `gorm:"column:secret_access_key"`
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
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// contains everything not in signup request but in the original user object
type LinkRequest struct {
	AwsCreds  AwsCreds  `gorm:"embedded;embeddedPrefix:aws_"`
	LLMConfig LLM       `gorm:"embedded;embeddedPrefix:openai_"`
	Portfolio Portfolio `gorm:"embedded;embeddedPrefix:portfolio_"`
	Git       Git       `gorm:"embedded;embeddedPrefix:git_"`
}
