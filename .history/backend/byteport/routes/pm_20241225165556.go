package routes

import "github.com/gin-gonic/gin"

func terminateInstance(c *gin.Context) {
	user := c.MustGet("user")
	project := c.MustGet("project")
	project.User = user;


}