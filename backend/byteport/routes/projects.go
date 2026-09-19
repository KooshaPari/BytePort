package routes

import (
	"byteport/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProjects(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	var projects []models.Project
	if err := models.DB.Where("owner = ?", user.UUID).Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}
	for _, project := range projects {
		err := project.AfterFind(models.DB)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, projects)
}
