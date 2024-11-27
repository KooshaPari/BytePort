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
		var user models.User
models.DB.Where("email = ?", c.Query("email")).First(&user)

		if lib.ValidatePass(c.Query("password"), user.Password) {
			user.EncryptedToken, _ = lib.GenerateToken(user)
			user.Password = "";
			c.JSON(http.StatusOK, gin.H{
				"message": "Success",
				"user": user,
			})
			
			
		} else{
			fmt.Println("Invalid Credentials.")
			c.JSON(http.StatusOK, gin.H{
				"message": "Failed",
			})
		}
		
	  })
	r.POST("/signup", func(c *gin.Context) {
		
		hash := lib.EncryptPass(c.Query("password"))
		
      newUser := models.User{
		UUID: uuid.NewString(),
		Name: c.Query("name"),
		Email: c.Query("email"),
		Password: string(hash),
	  }
	  // check for pre-existing user
	  var existingUser models.User
	  models.DB.Where("email = ?", c.Query("email")).First(&existingUser)
	  if existingUser.Email != "" {
	  models.DB.Create(&newUser)
	  newUser.Password = "";
	  newUser.EncryptedToken, _ = lib.GenerateToken(newUser)
	  c.JSON(http.StatusOK, newUser)


  })

	return r
}

func main() {
	models.ConnectDatabase()
	lib.InitAuthSystem()
	r := setupRouter()


	r.Run(":8080")
}
