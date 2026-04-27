package routes

import (
	"byteport/models"

	"github.com/gin-gonic/gin"
)

func handleGitHubCallback(c *gin.Context) {
	
	installationID := c.Query("installation_id") // Extract the installation ID

	if installationID == "" {
		c.JSON(400, gin.H{"error": "installation_id not provided"})
		return
	}

	// Save the installation ID to the database
	err := saveGitHubInstallation(userID, installationID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to save installation"})
		return
	}

	c.JSON(200, gin.H{
		"message": "GitHub App installed successfully",
		"installation_id": installationID,
	})
}

func saveGitHubInstallation(userID, installationID string) error {
	var user models.User;
	models.DB.Where("uuid = ?", userID).First(&user)
	var integration = models.Git{
		InstallID: installationID,
	}

	// Save the installation ID to the database
	result := models.DB.Save(&integration)
	return result.Error
}