package routes

import "github.com/gin-gonic/gin"

func TerminateInstance(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	project := c.MustGet("project").(models.User)
	project.User = user;



}