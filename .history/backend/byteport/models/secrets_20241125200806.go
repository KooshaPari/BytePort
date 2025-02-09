package models

// add owning user uuid
type Secrets struct {

    AppID      string `json:"app_id"`          // Global GitHub App ID
    PrivateKey string `json:"private_key"`     // Encrypted private key (PEM-encoded)
    APIBaseURL string `json:"api_base_url"`    // GitHub API base URL (e.g., https://api.github.com)

}