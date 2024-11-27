package routes

import (
	"byteport/lib"
	"byteport/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleCallback(c *gin.Context) {
	fmt.Println("Handling callback...")
	// printout full query string
	fmt.Println(c.Request.URL.RawQuery)
	code := c.Query("code")
    state := c.Query("state")

    // Validate state
    var user models.User
    if err := models.DB.Where("uuid = ?", state).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid state parameter"})
        return
    }

    // Exchange code for user access token
	secrets := models.GitSecret{}
	models.DB.First(&secrets)
    tokenUrl := "https://github.com/login/oauth/access_token"
    clientID,err := lib.DecryptSecret(secrets.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt client id"})
		return
	}
    clientSecret,err := lib.DecryptSecret(secrets.ClientSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt client secret"})
		return
	}

    payload := map[string]string{
        "client_id":     clientID,
        "client_secret": clientSecret,
        "code":          code,
    }
    payloadBytes, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", tokenUrl, bytes.NewBuffer(payloadBytes))
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch access token"})
        return
    }
    defer resp.Body.Close()

    var tokenResponse map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&tokenResponse)

    accessToken := tokenResponse["access_token"].(string)

    // Save the token securely for the user
    encryptedToken, err := lib.EncryptSecret(accessToken)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt access token"})
        return
    }
    user.Git.InstallID = encryptedToken
    models.DB.Save(&user)

    c.JSON(http.StatusOK, gin.H{"message": "GitHub linked successfully"})
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
	err= lib.ValidateGit(user)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Git credentials", "details": err.Error()})
		return
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
	if (notLinked)" {
		c.JSON(http.StatusOK, gin.H{"linked": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"linked": true})
}