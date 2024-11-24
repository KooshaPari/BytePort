package lib

import (
	"byteport/models"
	"time"

	"aidanwoods.dev/go-paseto"
)

func generateToken(user User) string {
	token := paseto.NewToken()
	token.SetAudience(user.Email)
	token.SetExpiration(time.Now().Add(time.Hour * 1))
	token.SetSubject("session")
	token.SetIssuer("BytePort")
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetString("user-id", "<uuid>")

	key := paseto.NewV4SymmetricKey()
	encryptedToken := token.V4Encrypt(key,nil)
	// Add to existing user entry
	var userDB User
	models.DB.db.Where("email = ?", user.Email).First(&userDB).
	
	return encryptedToken
}