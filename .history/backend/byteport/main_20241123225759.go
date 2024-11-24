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
	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.POST("/link", func(c *gin.Context) {
		var newData models.User 
		var user models.User

		if err := db.First(&user, "id = ?", userID).Error; err != nil {
			// handle error
		}
		if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
	if err := models.DB.Model(&user).Omit("Password", "Name", "Email").Updates(user).Error; err != nil {
    // handle error}
		
		// replace existing user obj with new posted one
		models.DB.Create(&user)
		c.JSON(http.StatusOK, user)

	}
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
