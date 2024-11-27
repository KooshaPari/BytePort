package models

// add owning user uuid
type GitSecret struct {

    GitAppID      string `json:"app_id"`          // Global GitHub App ID
    GitPrivateKey string `json:"private_key"`     // Encrypted private key (PEM-encoded)
    GitAPIBaseURL string `json:"api_base_url"`    // GitHub API base URL (e.g., https://api.github.com)

}
// insert git secret