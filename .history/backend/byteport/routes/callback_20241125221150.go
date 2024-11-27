package routes

import (
	"byteport/lib"
	"byteport/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleCallback(c *gin.Context) {
	
	state := c.Query("state")                // Unique token
	installationID := c.Query("installation_id") // GitHub installation ID

	if state == "" || installationID == "" {
		c.JSON(400, gin.H{"error": "Missing UUID or installation_id"})
		return
	}

	
	// Save the installation ID to the database
	err := saveGitHubInstallation(state, installationID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to save installation"})
		return
	}
	var user models.User 
	models.DB.Where("uuid = ?", state).First(&user)
	err= lib.ValidateGit(user.Git.InstallID)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Git credentials", "details": err.Error()})
		return
	}
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
func validateLink