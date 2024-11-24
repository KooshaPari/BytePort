package main

import (
	"byteport/models"
	"net/http"
	"byteport/lib"
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
	  
	r.POST("/signup", func(c *gin.Context) {
		hash := encryptPass(c.Query("password"))
		
      newUser := User{
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
