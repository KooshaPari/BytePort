package routes
func getInstances(c *gin.Context){
	var instances []models.Instance
	models.DB.Find(&instances)
	c.JSON(http.StatusOK, instances)
}