package main

import (
	"fmt"
	"net/http"
	"nvms/lib"
	"os"
	"strings"

	"aidanwoods.dev/go-paseto"
	spinhttp "github.com/fermyon/spin-go-sdk/http"
)
var (
    // Initialize during startup with key from environment/config
    serviceKey paseto.V4SymmetricKey
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		router := initRouter()
		router.ServeHTTP(w, r)
	})
   
}
func initRouter() *spinhttp.Router{
	router := spinhttp.NewRouter()
	router.GET("/", handle)
	return router;
}
func handle(w http.ResponseWriter, r *http.Request, params spinhttp.Params) {
    // Get key from environment/config during initialization
    keyHex := os.Getenv("SERVICE_KEY").encode()
    if keyHex == "" ||{
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
func ensureKeyExists() error {
    key := os.Getenv("SERVICE_KEY")
    if key == "" {
        return nil // Key already exists
    }

    // Generate and store a new key if not present
    newKey := lib.GenerateSymmetricKey()
    err:= os.Setenv("SERVICE_KEY", newKey)
	if err != nil {
		return err
	}
	return nil
}
func main() {
    // Main function is required for the Go program to run
    // You can add any initialization code here if needed
}