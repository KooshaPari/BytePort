package routes

func GetProjects(c *gin.Context){
	var projects []models.Project
	models.DB.Find(&projects).Where("owner = ?", c.MustGet("user").(models.User).UUID)
}