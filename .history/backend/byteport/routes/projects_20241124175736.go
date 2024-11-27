package routes

import (
	"byteport/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProjects(c *gin.Context){
	var projects []models.Project
	models.DB.Find(&projects).Where("owner = ?", c.MustGet("user").(models.User).UUID)
	c.JSON(http.StatusOK, projects)
}