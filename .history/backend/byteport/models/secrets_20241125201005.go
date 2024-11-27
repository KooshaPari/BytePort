package models

import (


	"gorm.io/gorm"
)

// add owning user uuid
type GitSecret struct {

    GitAppID      string `json:"app_id"`          // Global GitHub App ID
    GitPrivateKey string `json:"private_key"`     // Encrypted private key (PEM-encoded)
    GitAPIBaseURL string `json:"api_base_url"`    // GitHub API base URL (e.g., https://api.github.com)

}
func (g *GitSecret) BeforeSave(tx *gorm.DB) (err error) {
	if g.InstallationID != "" {
		g.InstallationID, err = lib.EncryptSecret(g.InstallationID)
		if err != nil {
			return err
		}
	}
	if g.Token != "" {
		g.Token, err = lib.EncryptSecret(g.Token)
		if err != nil {
			return err
		}
	}
	if g.SelectedRepo != "" {
		g.SelectedRepo, err = lib.EncryptSecret(g.SelectedRepo)
		if err != nil {
			return err
		}
	}
	return nil
}
func (g *GitSecret) AfterFind(tx *gorm.DB) (err error) {
	if g.InstallationID != "" {
		g.InstallationID, err = lib.DecryptSecret(g.InstallationID)
		if err != nil {
			return err
		}
	}
	if g.Token != "" {
		g.Token, err = lib.DecryptSecret(g.Token)
		if err != nil {
			return err
		}
	}
	if g.SelectedRepo != "" {
		g.SelectedRepo, err = lib.DecryptSecret(g.SelectedRepo)
		if err != nil {
			return err
		}
	}
	return nil
}