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
	var temp models.GitSecret
	result := models.DB.First(&temp) // Retrieve the first entry
	if result.Error != nil {
		if result.RowsAffected == 0 {
			fmt.Println("No entries found in git_secrets table.")
		} else {
			fmt.Printf("Error retrieving data from git_secrets table: %v\n", result.Error)
		}
		os.Exit(1)
	}

	r := setupRouter()
	

	r.Run(":8080")
}
