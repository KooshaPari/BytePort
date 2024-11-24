package main

import (
	"byteport/models"
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
	  
	r.POST("/signup", func(c *gin.Context) {
		pass = c.Query()
      newUser := User{
		Name: c.Query("name"),
		Email: c.Query("email"),
		Password: c.Query("password"),
	  }
	  models.DB.Create(&newUser)


  })

	return r
}

func main() {
	models.ConnectDatabase()
	r := setupRouter()


	r.Run(":8080")
}
