package lib

import (
	"fmt"
	"log"
	"net/http"
	"nvms/models"
	"os"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"

	"github.com/zalando/go-keyring"
)
const (
    keyringUser    = "BytePortUser"
    serviceKeyService = "NVMService"
)
var (
    
    serviceKey paseto.V4SymmetricKey
)
func GetSymmetricKey() (string, error) {
    return keyring.Get(serviceKeyService, keyringUser)
}
func ensureKeyExists(service, user string) error {
   key := os.Getenv("SERVICE_KEY")
    if key == "" {
        return nil // Key already exists
    }

    // Generate and store a new key if not present
    newKey := GenerateSymmetricKey()
    err:= os.Setenv("SERVICE_KEY", newKey)
	if err != nil {
		return err
	}
	return nil
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
	keyHex,err := GetSymmetricKey()
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

	keyHex, err := GetSymmetricKey()
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
func AuthMiddleware(w http.ResponseWriter, r *http.Request) {
    // Get key from environment/config during initialization
    keyHex := os.Getenv("SERVICE_KEY")
    if keyHex == "" || keyHex != "69438bb4a39c6e5ad86678846642a5c6f0b8a0299d467c40d674722e46805bcb"{
        http.Error(w, "Server configuration error", http.StatusInternalServerError)
        return
    }

    key, err := paseto.V4SymmetricKeyFromHex(keyHex)
    if err != nil {
        http.Error(w, "Server configuration error", http.StatusInternalServerError)
        return // Don't panic, return error response instead
    }
    serviceKey = key

    // Validate PASETO token
    authHeader, err := r.Cookie("Authorization")
    if err != nil {
        http.Error(w, "Unauthorized - No auth cookie found", http.StatusUnauthorized)
        return
    }

    authToken := authHeader.Value
    if authToken == "" {
        http.Error(w, "Unauthorized - Empty auth token", http.StatusUnauthorized)
        return
    }

    tokenString := strings.TrimPrefix(authToken, "Bearer ")
    parser := paseto.NewParser()
   
    token, err := parser.ParseV4Local(serviceKey, tokenString, nil)
    if err != nil {
        http.Error(w, "Unauthorized - Invalid token " + err.Error(), http.StatusUnauthorized)
        return
    }

    // Extract claims with error checking
    projectID, err := token.GetString("project-id")
    if err != nil {
        http.Error(w, "Invalid token claims", http.StatusBadRequest)
        return
    }

    userID, err := token.GetString("user-id")
    if err != nil {
        http.Error(w, "Invalid token claims", http.StatusBadRequest)
        return
    }

    fmt.Printf("Successfully authenticated user %s for project %s\n", userID, projectID)
    w.WriteHeader(http.StatusOK)
}
