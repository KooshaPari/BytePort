package models

// add owning user uuid
type GitSecret struct {

   ID         int    `gorm:"primaryKey"`
	PrivateKey string `gorm:"column:private_key"` // Encrypted private key
	AppID      string `gorm:"column:app_id"`      // GitHub App ID
	ClientID   string `gorm:"column:client_id"`

}
