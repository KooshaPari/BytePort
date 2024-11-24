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
	openAICreds struct{
		apiKey string
	}
	portfolio struct{
		rootEndpoint string
		apiKey string
	}
	git struct{
		repoUrl string
		authMethod string
		authKey string
		targetDirectory string
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
      newUser := User{
		Name: c.Query("name"),
		Email: c.Query("email"),
		Password: c.Query("password"),
	  }


  })

	return r
}

func main() {
	byteport.ConnectDatabase()
	r := setupRouter()


	r.Run(":8080")
}
