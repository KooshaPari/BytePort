package main

import (
	"byteport/lib"
	"byteport/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var db = make(map[string]string)


func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	  r.POST("/login", func(c *gin.Context) {
		var user models.User
models.DB.Where("email = ?", c.Query("email")).First(&user)

		if lib.ValidatePass(c.Query("password"), user.Password) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Success",
			})
			user.Enc
		} else{
			fmt.Println("Invalid Credentials.")
			c.JSON(http.StatusOK, gin.H{
				"message": "Failed",
			})
		}
		
	  })
	r.POST("/signup", func(c *gin.Context) {
		hash := lib.EncryptPass(c.Query("password"))
		
      newUser := lib.User{
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
