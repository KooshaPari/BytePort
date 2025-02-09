package lib

import (
	"errors"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/zalando/go-keyring"
)
const (
    keyringService = "BytePortService"
    keyringUser    = "BytePortUser"
)

func storeSymmetricKey(key string) error {
    return keyring.Set(keyringService, keyringUser, key)
}

func getSymmetricKey() (string, error) {
    return keyring.Get(keyringService, keyringUser)
}
func InitAuthSystem() {
	keyHex := generateSymmetricKey()
	storeSymmetricKey(keyHex)

}
func generateSymmetricKey() string {
    key := paseto.NewV4SymmetricKey()
    return key.ExportHex()
}

func generateToken(user User) (string,error) {
	token := paseto.NewToken()
	token.SetAudience(user.Email)
	token.SetExpiration(time.Now().Add(time.Hour * 1))
	token.SetSubject("session")
	token.SetIssuer("BytePort")
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetString("user-id", user.ID.String())

	if keyHex == "" {
        return "", errors.New("symmetric key not set")
    }
	key, err := paseto.V4SymmetricKeyFromHex(keyHex)
    if err != nil {
        return "", err
    }

	encryptedToken := token.V4Encrypt(key,nil)
	

	
	return encryptedToken,nil
}