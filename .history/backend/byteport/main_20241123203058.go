package main

import (
	"byteport/lib"
	"byteport/models"
	"net/http"

	"github.com/gin-gonic/gin"
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
		} else{
			c.JSON(http.StatusOK, gin.H{
				
		}
	  })
	r.POST("/signup", func(c *gin.Context) {
		hash := lib.EncryptPass(c.Query("password"))
		
      newUser := lib.User{
		Name: c.Query("name"),
		Email: c.Query("email"),
		Password: string(hash),
	  }
	  models.DB.Create(&newUser)
	  newUser.Password = "";
	  c.JSON(http.StatusOK, newUser)


  })

	return r
}

func main() {
	models.ConnectDatabase()
	
	r := setupRouter()


	r.Run(":8080")
}
