package models

// add owning user uuid
type GitSecret struct {

    GitAppID      string `json:"app_id"`          // Global GitHub App ID
    GitPrivateKey string `json:"private_key"`     // Encrypted private key (PEM-encoded)
    ClientID string `json:"client_id"`    	// GitHub App client ID

}
