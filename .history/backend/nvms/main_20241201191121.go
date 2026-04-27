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
    keyHex := os.Getenv("SERVICE_KEY")
    key, err := paseto.V4SymmetricKeyFromHex(keyHex)
    if err != nil {
        panic("Invalid service key")
    }
    serviceKey = key
	fmt.P
    // Validate PASETO token
    authHeader,err := r.Cookie("Authorization")
    if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
        return
    }
	authToken := authHeader.Value;
    tokenString := strings.TrimPrefix(authToken, "Bearer ")
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
    
    fmt.Println("Successfully authenticated user", userID, "for project", projectID);
	return;
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