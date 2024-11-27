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
	lib.InitAuthSystem()
	// Read the private key from file
	privateKey, err := os.ReadFile("byteport-ghkey.pem")
	if err != nil {
		fmt.Printf("Failed to read private key file: %v\n", err)
		return
	}

	// Encrypt the private key
	encryptedPrivateKey, err := lib.EncryptSecret(string(privateKey))
	if err != nil {
		fmt.Printf("Failed to encrypt private key: %v\n", err)
		return
	}
	encryptedClientID, err := lib.EncryptSecret("Iv23lippK4CVHY7jLdFv")
	if err != nil {
		fmt.Printf("Failed to encrypt client key: %v\n", err)
		return
	}
	encryptedAppID, err := lib.EncryptSecret("1069682")
	if err != nil {
		fmt.Printf("Failed to encrypt app key: %v\n", err)
	var GitData = models.GitSecret{
		ClientID:     encryptedClientID,
		AppID:    encryptedAppID,
		PrivateKey: encryptedPrivateKey,
	}
	
	models.DB.Where("app_id = ?", encryptedAppID).First(&GitData)
	models.DB.Create(&GitData)
	}
	//del from users
	//models.DB.Exec("DELETE FROM users")
	r := setupRouter()
	

	r.Run(":8080")
}
