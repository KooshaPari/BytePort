package main

import (
	"byteport/lib"
	"byteport/models"
	"byteport/routes"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
		
    }))
	protected := r.Group("/")
	protected.Use(lib.AuthMiddleware())
	{
	protected.POST("/link", routes.LinkHandler)
	protected.POST("/authenticate", routes.Authenticate)
	protected.GET("/instances", routes.GetInstances)
	
	}
	r.POST("/login", routes.Login)
	r.POST("/signup", routes.Signup)

	return r
}

func main() {
	models.ConnectDatabase()
	models.DB.Exec("DELETE FROM users")
	lib.InitAuthSystem()
	r := setupRouter()


	r.Run(":8080")
}
