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
        AllowOrigins: []string{
			"http://localhost:5173",
			"http://0.0.0.0:5173",
			"http://tauri.localhost",
			"http://tauri.0.0.0.0:5173",
			"http://localhost:8081",
			"http://0.0.0.0:8081",
			"http://10.0.2.2:5173",
			"http://10.0.2.2:8081",
			// Add other needed origins
		},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        AllowWildcard:    true,  // Enable wildcard matching
        MaxAge:           12 * time.Hour,
    }))
	protected := r.Group("/")
	protected.Use(lib.AuthMiddleware())
	{

	protected.GET("/link", routes.LinkHandler)
	protected.POST("/link", routes.ValidateLink)
	protected.GET("/authenticate", routes.Authenticate)
	protected.GET("/instances", routes.GetInstances)
	protected.GET("/projects", routes.GetProjects)
	protected.GET("/api/github/repositories", routes.RetrieveRepositories)
	protected.POST("/deploy", routes.DeployProject)
	//protected.GET("/github/status", routes.GitHubStatusHandler)
	}
	r.POST("/login", routes.Login)
	r.POST("/signup", routes.Signup)
	r.GET("/api/github/callback", routes.HandleCallback)
 
	// gh webhook at /api/github/auth/webhook
	

	return r
}

func main() {
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
	//models.DB.Exec("Delete from projects")
	r := setupRouter()
	//models.DB.Exec("Delete from users")

	go lib.StartTokenRefreshJob()
	if err := r.Run("0.0.0.0:8081"); err != nil {
        fmt.Printf("Error starting server: %v\n", err)
        os.Exit(1)
    }
}
