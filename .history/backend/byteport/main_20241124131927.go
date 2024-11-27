package main

import (
	"byteport/lib"
	"byteport/models"
	"byteport/routes"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var db = make(map[string]string)


func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	
	r := gin.Default()
	r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:5173"}, // Your frontend origin
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
	protected.POST("/authenticate", {
		userInterface, exists := c.Get("user")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        user := userInterface.(models.User)

        c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"UUID": user.UUID,
			"EncryptedToken": user.EncryptedToken,
		})
		
	})
	}
	  
	r.POST("/login", routes.Login)
	r.POST("/signup", routes.Signup)


	return r
}

func main() {
	models.ConnectDatabase()
	// clear user table
	models.DB.Exec("DELETE FROM users")
	lib.InitAuthSystem()
	r := setupRouter()


	r.Run(":8080")
}
