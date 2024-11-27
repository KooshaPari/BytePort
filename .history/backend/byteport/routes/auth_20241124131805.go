package routes

import (
	"byteport/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func linkHandler(c *gin.Context) {
	// update user obj with new posted one (sans email pass name)
	user := c.MustGet("user").(models.User)
	var req models.LinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.AwsCreds = req.AwsCreds
	user.OpenAICreds = req.OpenAICreds
	user.Portfolio = req.Portfolio
	user.Git = req.Git
	models.DB.Save(&user)
	c.JSON(http.StatusOK, user)
	
}
func login()