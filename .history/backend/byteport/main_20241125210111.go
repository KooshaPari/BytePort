package main

import (
	"byteport/lib"
	"byteport/models"
	"byteport/routes"
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
		
    }))
	protected := r.Group("/")
	protected.Use(lib.AuthMiddleware())
	{
	protected.POST("/link", routes.LinkHandler)
	protected.GET("/authenticate", routes.Authenticate)
	protected.GET("/instances", routes.GetInstances)
	protected.GET("/projects", routes.GetProjects)
	}
	r.POST("/login", routes.Login)
	r.POST("/signup", routes.Signup)

	return r
}

func main() {
	os.Remove("ENCRYPTION_KEY")
	err := lib.InitializeEncryptionKey()
	if err != nil {
		fmt.Printf("Error initializing encryption key: %v\n", err)
		os.Exit(1)
	}
	
	models.ConnectDatabase()
	err = lib.InitAuthSystem()
	if err != nil {
		fmt.Printf("Error initializing auth system: %v\n", err)
		os.Exit(1)
	}

	// Read the private key from file
	privateKey, err := os.ReadFile("byteport-ghkey.pem")
	if err != nil {
		fmt.Printf("Failed to read private key file: %v\n", err)
		os.Exit(1)
	}

	// Encrypt sensitive GitHub App data
	encryptedPrivateKey, err := lib.EncryptSecret(string(privateKey))
	if err != nil {
		fmt.Printf("Failed to encrypt private key: %v\n", err)
		os.Exit(1)
	}

	clientID := "Iv23lippK4CVHY7jLdFv" 
	encryptedClientID, err := lib.EncryptSecret(clientID)
	if err != nil {
		fmt.Printf("Failed to encrypt Client ID: %v\n", err)
		os.Exit(1)
	}

	appID := "1069682" 
	encryptedAppID, err := lib.EncryptSecret(appID)
	if err != nil {
		fmt.Printf("Failed to encrypt App ID: %v\n", err)
		os.Exit(1)
	}

	// Create GitSecret object
	var GitData = models.GitSecret{
		ClientID:   encryptedClientID,
		AppID:      encryptedAppID,
		PrivateKey: encryptedPrivateKey,
	}

	// Store the encrypted data in the database
	err = models.DB.Create(&GitData).Error
	if err != nil {
		fmt.Printf("Failed to save GitSecret to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("GitHub App data stored securely in the database.")
	//del from users
	models.DB.Exec("DELETE FROM git_secrets")
	r := setupRouter()
	

	r.Run(":8080")
}
