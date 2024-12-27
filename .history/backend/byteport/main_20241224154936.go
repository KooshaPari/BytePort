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
    "http://localhost:8081",
    "http://0.0.0.0:8081",
    // Add other needed origins
},
Copy
Retry


KP
my svelte frontend is unable to fetch teh url, most likely bc tauri isnt allowing it


Yes, we need to update the Tauri config to allow these network requests. In your tauri.conf.json, update the allowlist section under app:

json

Copy
{
  "app": {
    "security": {
      "csp": null
    },
    "allowlist": {
      "http": {
        "all": true,
        "request": true,
        "scope": [
          "http://0.0.0.0:8081/*",
          "http://localhost:8081/*",
          "http://localhost:5173/*",
          "http://127.0.0.1:8081/*"
        ]
      }
    }
  }
}
Also, for working with mobile builds, you'll likely need to use your machine's actual IP address. You can add that to both:

The allowlist scope in tauri.conf.json:
json

Copy
"scope": [
  "http://0.0.0.0:8081/*",
  "http://192.168.1.x:8081/*"  // Your actual IP
]
The CORS configuration in your Gin server:
go

Copy
AllowOrigins: []string{
    "http://localhost:5173",
    "http://0.0.0.0:5173",
    "http://192.168.1.x:5173",  // Your actual IP
},
Copy
Retry


KP
tauri now uses a new system that doesn thave allowlist


For Tauri 2.0, the network permissions are configured differently. In your tauri.conf.json, add the security permissions for networking:

json

Copy
{
  "app": {
    "security": {
      "csp": null,
      "permissions": {
        "network": {
          "hosts": [
            "http://0.0.0.0:8081",
            "http://localhost:8081",
            "http://127.0.0.1:8081",
            "http://192.168.1.*:8081"
          ],
          "protocols": ["http", "https"]
        }
      }
    }
  }
}
This new permission model in Tauri 2.0 replaces the older allowlist system. The permissions are defined under app.security.permissions and allow you to specify network hosts and protocols that your app can connect to.

Copy
Retry


Claude can make mistakes. Please double-check responses.
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
	models.DB.Exec("Delete from projects")
	r := setupRouter()
	//models.DB.Exec("Delete from users")

	go lib.StartTokenRefreshJob()
	if err := r.Run("0.0.0.0:8081"); err != nil {
        fmt.Printf("Error starting server: %v\n", err)
        os.Exit(1)
    }
}
