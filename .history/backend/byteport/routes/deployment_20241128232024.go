package routes

import "github.com/gin-gonic/gin"

func DeployProject(c *gin.Context) (error){
	// add project to db after compiling to obj;
	// get project from request
	// compile project
	var newProject models.Project
	
}