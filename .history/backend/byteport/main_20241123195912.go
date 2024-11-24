package main

import (
	"byteport/models"
	"log"
	"net/http"

	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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
	r.Use(handlerMiddleWare(authMiddleware))
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
	  newUser.Password = "";
	  c.JSON(http.StatusOK, newUser)


  })

	return r
}

func main() {
	models.ConnectDatabase()
	authMiddleware, err := jwt.New(initParams())
  if err != nil {
    log.Fatal("JWT Error:" + err.Error())
  }
	r := setupRouter()


	r.Run(":8080")
}
