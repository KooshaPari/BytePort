package routes

import "github.com/gin-gonic/gin"

func TerminateInstance(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	}
	project.User = user;



}