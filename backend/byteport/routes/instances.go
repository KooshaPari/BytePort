package routes

import (
	"byteport/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetInstances(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	var instances []models.Instance
	if err := models.DB.Where("owner = ?", user.UUID).Find(&instances).Error; err != nil {
		respondInternalError(c, "Failed to fetch instances")
		return
	}
	c.JSON(http.StatusOK, instances)
}
