package routes

import (
	"byteport/lib"
	"byteport/models"
	"fmt"
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
func ValidateLink(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	var req models.User
	// Bind JSON to the request model
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}
	var err error
	err = lib.ValidateAWSCredentials(req.AwsCreds.AccessKeyID,req.AwsCreds.SecretAccessKey)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate AWS credentials", "details": err.Error()})
		return
	}


	
	err= lib.ValidateGit(user.Git.InstallID)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Git credentials", "details": err.Error()})
		return
	}
	err=lib.ValidateOpenAICredentials(req.OpenAICreds.APIKey)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate OAI credentials", "details": err.Error()})
		return
	}
	err=lib.ValidatePortfolioAPI(req.Portfolio.RootEndpoint,req.Portfolio.APIKey)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Portfolio credentials", "details": err.Error()})
		//return
	}
	

	encryptedPortfolioAPIKey, err := lib.EncryptSecret(req.Portfolio.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Portfolio API Key", "details": err.Error()})
		return
	}

	encryptedInstallID, err := lib.EncryptSecret(req.Git.InstallID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Git InstallID", "details": err.Error()})
		return
	}

	encryptedAccessKeyID, err := lib.EncryptSecret(req.AwsCreds.AccessKeyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt AWS Access Key ID", "details": err.Error()})
		return
	}

	encryptedSecretAccessKey, err := lib.EncryptSecret(req.AwsCreds.SecretAccessKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt AWS Secret Access Key", "details": err.Error()})
		return
	}

	encryptedApiKey, err := lib.EncryptSecret(req.OpenAICreds.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt OpenAI API Key", "details": err.Error()})
		return
	}

	encryptedPortfolioURL, err := lib.EncryptSecret(req.Portfolio.RootEndpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Portfolio Root Endpoint", "details": err.Error()})
		return
	}

	// Update user fields
	user.Git = models.Git{
		InstallID: encryptedInstallID,
	}

	user.AwsCreds = models.AwsCreds{
		AccessKeyID:     encryptedAccessKeyID,
		SecretAccessKey: encryptedSecretAccessKey,
	}

	user.OpenAICreds = models.OpenAICreds{
		APIKey: encryptedApiKey,
	}

	user.Portfolio = models.Portfolio{
		RootEndpoint: encryptedPortfolioURL,
		APIKey:       encryptedPortfolioAPIKey,
	}

	// Save the updated user
	if err := models.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user", "details": err.Error()})
		return
	}
}
func GitHubStatusHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	// Check if the user has a GitHub installation ID
	fmt.Println("ID: ",user.Git.InstallID)
	if user.Git.InstallID == "" {
		c.JSON(http.StatusOK, gin.H{"linked": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"linked": true})
}