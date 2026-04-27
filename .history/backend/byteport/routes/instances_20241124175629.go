package routes

import (
	"byteport/models"
	"net/http"

	"github.com/gin-gonic/gin"
)
func getInstances(c *gin.Context){
	var instances []models.Instance
	models.DB.Find(&instances)
	c.JSON(http.StatusOK, instances)
}