package routes

import (
	"byteport/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetInstances(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user, ok := userVal.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	var instances []models.Instance
	if err := models.DB.Where("owner = ?", user.UUID).Find(&instances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch instances"})
		return
	}
	c.JSON(http.StatusOK, instances)
}
