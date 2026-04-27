package routes

func terminateInstance(c *gin.Context) {
	user := c.MustGet("user")
	project := c.MustGet("project")
	project.user = user;
	

}