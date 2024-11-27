package main

import (
	"byteport/lib"
	"byteport/models"
	"fmt"
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
	protected.Use(AuthMiddleware())
	{
	protected.POST("/link", linkHandler)
	protected.POST("/authenticate", func(c *gin.Context){
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

	  r.POST("/login", func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		if err := models.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Failed, user not found"})
			return
		}

		if lib.ValidatePass(req.Password, user.Password) {
			user.EncryptedToken, _ = lib.GenerateToken(user)
			user.Password = ""
			c.JSON(http.StatusOK, gin.H{
				"message": "Success",
				"user":    user,
			})
		} else {
			fmt.Println("Invalid Credentials.")
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Failed, invalid credentials",
			})
		}
	})

	// Updated /signup handler to process JSON
	r.POST("/signup", func(c *gin.Context) {
		var req models.SignupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash := lib.EncryptPass(req.Password)

		// Check for pre-existing user
		var existingUser models.User
		if err := models.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			// User already exists
			c.JSON(http.StatusConflict, gin.H{
				"message": "Failed, User Already Exists",
			})
			return
		}

		newUser := models.User{
			UUID:     uuid.NewString(),
			Name:     req.Name,
			Email:    req.Email,
			Password: string(hash),
		}

		// Create the new user
		if err := models.DB.Create(&newUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to create user",
				"error":   err.Error(),
			})
			return
		}

		newUser.Password = ""
		newUser.EncryptedToken, _ = lib.GenerateToken(newUser)
		c.JSON(http.StatusCreated, newUser)
	})

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
