package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var db = make(map[string]string)
type User struct{
	Name string
	Email string
	Password string
	awsCreds struct{
		accessKeyId string
		secretAccessKey string
	}
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	  
	r.GET("/signup", func(c *gin.Context) {
      
  })

	return r
}

func main() {
	r := setupRouter()


	r.Run(":8080")
}
