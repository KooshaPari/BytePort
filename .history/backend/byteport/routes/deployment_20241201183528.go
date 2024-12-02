package routes

import (
	"byteport/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeployProject(c *gin.Context){
	// add project to db after compiling to obj;
	// get project from request
	// compile project
	var newProject models.Project
	if err := c.ShouldBindJSON(&newProject); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	}
	models.DB.Create(&newProject)
	fmt.Println("Deploying project: ", newProject)
	c.Request
	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"project": newProject,
	})

}