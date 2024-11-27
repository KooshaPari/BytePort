package main

import (
	"byteport/lib"
	"byteport/models"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var db = make(map[string]string)
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from headers
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
            c.Abort()
            return
        }

        // Validate token and get user
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        valid, token, err := lib.ValidateToken(tokenString)
        if err != nil || !valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        userID,_ := token.GetString("user-id")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
            c.Abort()
            return
        }

        // Retrieve user from database
        var user models.User
        if err := models.DB.Where("uuid = ?", userID).First(&user).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
            c.Abort()
            return
        }

        // Set user in context
        c.Set("user", user)
        c.Next()
    }
}
func linkHandler(c *gin.Context) {
	// update user obj with new posted one (sans email pass name)
	user := c.MustGet("user").(models.User)
	var req models.LinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.AwsCreds = req.AwsCreds
	user.OpenAICreds = req.OpenAICreds
	user.Portfolio = req.Portfolio
	user.Git = req.Git
	models.DB.Save(&user)
	c.JSON(http.StatusOK, user)
	
}
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
			"message": "Success",}}
		// validate toke
	})
	}
	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	/*r.POST("/link", func(c *gin.Context) {
		var user models.User 
		
		
		// replace existing user obj with new posted one
		models.DB.Create(&user)
		c.JSON(http.StatusOK, user)

	})*/
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
