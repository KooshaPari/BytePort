package lib

import (
	"byteport/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/gin-gonic/gin"
	"github.com/zalando/go-keyring"
)

const (
	tokenKeyService   = "BytePortTokenKeyService"
	secretsKeyService = "BytePortSecretsKeyService"
	keyringUser       = "BytePortUser"
	serviceKeyService = "NVMService"
)

// keyringTimeout bounds every keychain call.
//
// On macOS zalando/go-keyring shells out to /usr/bin/security with no timeout
// and no way to point it elsewhere (keyring_darwin.go pins
// execPathKeychain = "/usr/bin/security"). When the login keychain is locked,
// or the SecurityAgent is not responding, that call blocks indefinitely. Since
// InitAuthSystem runs before the router is mounted, the server prints
// "Connected to SQLite database" and then never binds a port: a supervisor sees
// a live process holding no port, forever, with no error to act on.
//
// Observed on this host: the process sat in state S for over 2.5 minutes with
// no listener, its only child being a blocked
// `/usr/bin/security find-generic-password`.
//
// Bounding these calls turns an indefinite hang into a clear startup error.
const keyringTimeout = 10 * time.Second

// keyringGet is keyring.Get with a deadline.
func keyringGet(service, user string) (string, error) {
	type result struct {
		secret string
		err    error
	}
	done := make(chan result, 1) // buffered: a slow goroutine must not leak
	go func() {
		secret, err := keyring.Get(service, user)
		done <- result{secret, err}
	}()
	select {
	case r := <-done:
		return r.secret, r.err
	case <-time.After(keyringTimeout):
		return "", fmt.Errorf(
			"keychain read for service %q timed out after %s; is the login keychain unlocked?",
			service, keyringTimeout)
	}
}

// keyringSet is keyring.Set with a deadline.
func keyringSet(service, user, secret string) error {
	done := make(chan error, 1)
	go func() {
		done <- keyring.Set(service, user, secret)
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(keyringTimeout):
		return fmt.Errorf(
			"keychain write for service %q timed out after %s; is the login keychain unlocked?",
			service, keyringTimeout)
	}
}

func getSymmetricKey() (string, error) {
	return keyringGet(tokenKeyService, keyringUser)
}
func ensureKeyExists(service, user string) error {
	_, err := keyringGet(service, user)
	if err == nil {
		log.Printf("Key already exists: %s\n", service)
		return nil // Key already exists
	}

	// Generate and store a new key if not present
	newKey := generateSymmetricKey()
	if service == serviceKeyService {
		log.Println("Setting service key")
		err = os.Setenv("SERVICE_KEY", newKey)
		if err != nil {
			return err
		}
	}
	return keyringSet(service, user, newKey)
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
	err = ensureKeyExists(serviceKeyService, keyringUser)
	if err != nil {
		return fmt.Errorf("failed to initialize service key: %w", err)
	}
	log.Println("Auth system initialized with separate keys for tokens and secrets.")
	return nil

}
func generateSymmetricKey() string {
	key := paseto.NewV4SymmetricKey()
	return key.ExportHex()
}

func GenerateToken(user models.User) (string, error) {
	token := paseto.NewToken()
	token.SetAudience(user.Email)
	token.SetExpiration(time.Now().Add(time.Hour * 1))
	token.SetSubject("session")
	token.SetIssuer("BytePort")
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetString("user-id", user.UUID)
	keyHex, err := getSymmetricKey()
	if err != nil {
		log.Fatal(err)
	}
	key, err := paseto.V4SymmetricKeyFromHex(keyHex)
	if err != nil {
		return "", err
	}

	encryptedToken := token.V4Encrypt(key, nil)

	return encryptedToken, nil
}
func GenerateNVMSToken(project models.Project) (string, error) {
	token := paseto.NewToken()
	token.SetAudience(serviceKeyService)
	token.SetExpiration(time.Now().Add(time.Minute * 10))
	token.SetSubject("deployment")
	token.SetIssuer("BytePort")
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetString("user-id", project.User.UUID)
	token.SetString("project-id", project.UUID)
	keyHex, err := keyringGet(serviceKeyService, keyringUser)
	if err != nil {
		log.Fatal(err)
	}

	key, err := paseto.V4SymmetricKeyFromHex(keyHex)
	if err != nil {
		return "", err
	}

	encryptedToken := token.V4Encrypt(key, nil)

	return encryptedToken, nil
}
func AuthenticateRequest(encryptedToken string) (*models.User, error) {
	valid, token, err := ValidateToken(encryptedToken)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	// Extract user ID from the token
	userID, err := token.GetString("user-id")
	if err != nil {
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
	user.Password = ""

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

	// Service tokens are encrypted with the (serviceKeyService, keyringUser)
	// key, NOT the session token key. Previously this read getSymmetricKey()
	// which returns the session token key, so GenerateNVMSToken and
	// ValidateServiceToken could never agree on a MAC: every "validate"
	// attempt returned "bad message authentication code". GenerateNVMSToken is
	// not currently called from production, so the bug was latent; fix the
	// symmetry so the round-trip works if/when the NVMS integration ships.
	keyHex, err := keyringGet(serviceKeyService, keyringUser)
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
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Extract token from headers
		authToken, _ := c.Cookie("authToken")
		if authToken == "" {
			log.Println("Auth Token Missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		// Validate token and get user
		tokenString := strings.TrimPrefix(authToken, "Bearer ")
		valid, token, err := ValidateToken(tokenString)
		if err != nil || !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		userID, _ := token.GetString("user-id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			c.Abort()
			return
		}

		// Retrieve user from database
		var user models.User
		if err := models.DB.Where("uuid = ?", userID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Next()
	}
}
