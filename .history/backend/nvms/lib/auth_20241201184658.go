package lib

import (
	"fmt"
	"log"
	"nvms/models"
	"time"

	"aidanwoods.dev/go-paseto"

	"github.com/zalando/go-keyring"
)
const (
    keyringUser    = "BytePortUser"
    serviceKeyService = "NVMService"
)

func GetSymmetricKey() (string, error) {
    return keyring.Get(serviceKeyService, keyringUser)
}
func ensureKeyExists(service, user string) error {
    _, err := keyring.Get(service, user)
    if err == nil {
        return nil // Key already exists
    }

    // Generate and store a new key if not present
    newKey := GenerateSymmetricKey()
    return keyring.Set(service, user, newKey)
}


func InitAuthSystem() error {
    err := ensureKeyExists(serviceKeyService, keyringUser)
    if err != nil {
        return fmt.Errorf("failed to initialize token key: %w", err)
    }

    // Initialize service key
    err = ensureKeyExists(serviceKeyService, keyringUser)
    if err != nil {
        return fmt.Errorf("failed to initialize secrets key: %w", err)
    }

    log.Println("Auth system initialized with separate keys for tokens and secrets.")
    return nil
	

}
func GenerateSymmetricKey() string {
    key := paseto.NewV4SymmetricKey()
    return key.ExportHex()
}

func GenerateNVMSToken(project models.Project) (string,error) {
	token := paseto.NewToken()
	token.SetAudience(serviceKeyService)
	token.SetExpiration(time.Now().Add(time.Minute * 10))
	token.SetSubject("deployment")
	token.SetIssuer("BytePort")
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetString("user-id", project.User.UUID)
    token.SetString("project-id", project.UUID)
	keyHex,err := getSymmetricKey()
	if(err != nil){
		log.Fatal(err)
	}
	key, err := paseto.V4SymmetricKeyFromHex(keyHex)
    if err != nil {
        return "", err
    }

	encryptedToken := token.V4Encrypt(key,nil)
	

	
	return encryptedToken,nil
}

func ValidateServiceToken(encryptedToken string) (bool, *paseto.Token, error) {

	keyHex, err := getSymmetricKey()
	if err != nil {
		return false, nil, err
	}

	key, err := paseto.V4SymmetricKeyFromHex(keyHex)
	if err != nil {
		return false, nil, err
	}
    
	parser := paseto.NewParser()
    parser.AddRule(paseto.ForAudience(serviceKeyService))
    parser.AddRule(paseto.NotExpired())
    
	token, err := parser.ParseV4Local(key, encryptedToken, nil)
	if err != nil {
		return false, nil, err
	}
	



	return true, token, nil
}