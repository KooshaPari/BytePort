package lib

import (
	"byteport/models"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/zalando/go-keyring"
)
const (
	tokenKeyService   = "BytePortTokenKeyService"
    secretsKeyService = "BytePortSecretsKeyService"
    keyringUser    = "BytePortUser"
    serviceKeyService = "NVMService"
)

func getSymmetricKey() (string, error) {
    return keyring.Get(tokenKeyService, keyringUser)
}
func ensureKeyExists(service, user string) error {
    _, err := keyring.Get(service, user)
    if err == nil {
        return nil // Key already exists
    }

    // Generate and store a new key if not present
    newKey := generateSymmetricKey()
    return keyring.Set(service, user, newKey)
}

func getTokenKey() (string, error) {
    return keyring.Get(tokenKeyService, keyringUser)
}

func getSecretsKey() (string, error) {
    return keyring.Get(secretsKeyService, keyringUser)
}
func InitAuthSystem() error {
    err := ensureKeyExists(tokenKeyService, keyringUser)
    if err != nil {
        return fmt.Errorf("failed to initialize token key: %w", err)
    }

    // Initialize secrets key
    err = ensureKeyExists(secretsKeyService, keyringUser)
    if err != nil {
        return fmt.Errorf("failed to initialize secrets key: %w", err)
    }

    log.Println("Auth system initialized with separate keys for tokens and secrets.")
    return nil
	

}
func generateSymmetricKey() string {
    key := paseto.NewV4SymmetricKey()
    return key.ExportHex()
}

func GenerateToken(user models.User) (string,error) {
	token := paseto.NewToken()
	token.SetAudience(user.Email)
	token.SetExpiration(time.Now().Add(time.Hour * 1))
	token.SetSubject("session")
	token.SetIssuer("BytePort")
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetString("user-id", user.UUID)
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
func AuthenticateRequest(encryptedToken string) (*models.User, error) {
	valid, token, err := ValidateToken(encryptedToken)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	// Extract user ID from the token
	userID, err := token.GetString("user-id")
	if(err != nil){
		log.Fatal(err)
	}
	if userID == "" {
		return nil, fmt.Errorf("user-id claim missing in token")
	}

	// Retrieve the user from the database
	var user models.User
	err = models.DB.Where("uuid = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	user.Password = "";

	return &user, nil
}
func ValidateToken(encryptedToken string) (bool, *paseto.Token, error) {

	keyHex, err := getSymmetricKey()
	if err != nil {
		return false, nil, err
	}

	key, err := paseto.V4SymmetricKeyFromHex(keyHex)
	if err != nil {
		return false, nil, err
	}

	parser := paseto.NewParser()

	token, err := parser.ParseV4Local(key, encryptedToken, nil)
	if err != nil {
		return false, nil, err
	}
	



	return true, token, nil
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
// SpinAuthMiddleware middleware for Spin service
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
            c.Abort()
            return
        }

        // Validate service token
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        valid, token, err := ValidateServiceToken(tokenString)
        if err != nil || !valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired service token"})
            c.Abort()
            return
        }

        // Extract project and user context
        projectID, err := token.GetString("project-id")
        userID, err := token.GetString("user-id")
        if projectID == "" || userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            c.Abort()
            return
        }

        // Set context for handler
        c.Set("projectID", projectID)
        c.Set("userID", userID)
        c.Next()
    }
}