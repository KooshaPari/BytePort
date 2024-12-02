package main

import (
	"net/http"
	"os"
	"strings"

	"aidanwoods.dev/go-paseto"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)
var (
    // Initialize during startup with key from environment/config
    serviceKey paseto.V4SymmetricKey
)

func init() {
    // Get key from environment/config during initialization
    keyHex := os.Getenv("SERVICE_KEY")
    key, err := paseto.V4SymmetricKeyFromHex(keyHex)
    if err != nil {
        panic("Invalid service key")
    }
    serviceKey = key
    
    spinhttp.Handle(handle)
}

func handle(w http.ResponseWriter, r *http.Request) {
    // Validate PASETO token
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    parser := paseto.NewParser()
    parser.AddRule(paseto.ForAudience("spin-deployment-service"))
    parser.AddRule(paseto.NotExpired())
    
    token, err := parser.ParseV4Local(serviceKey, tokenString, nil)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    // Extract claims
    projectID, _ := token.GetString("project-id")
    userID, _ := token.GetString("user-id")
    
    // Process deployment...
}
func ensureKeyExists() error {
    _, err := os.Getenv("SERVICE_KEY")
    if err == nil {
        return nil // Key already exists
    }

    // Generate and store a new key if not present
    newKey := generateSymmetricKey()
    err:= os.Setenv("SERVICE_KEY", newKey)
}