package main

import (
	"byteport/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
		pass := []byte(c.Query("password"))
		hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }
      newUser := User{
		Name: c.Query("name"),
		Email: c.Query("email"),
		Password: string(hash),
	  }
	  models.DB.Create(&newUser)
	  newUser.Password = nil;
	  c.JSON(http.StatusOK, newUser)


  })

	return r
}

func main() {
	models.ConnectDatabase()
	r := setupRouter()


	r.Run(":8080")
}
