package routes

import (
	"byteport/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeployProject(c *gin.Context) (error){
	// add project to db after compiling to obj;
	// get project from request
	// compile project
	var newProject models.Project
	if err := c.ShouldBindJSON(&newProject); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Deploying project: ", newProject)
	models.DB.Create(&newProject)
}