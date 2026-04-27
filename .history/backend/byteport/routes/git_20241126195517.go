package routes

import (
	"byteport/lib"
	"byteport/models"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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
      payloadBytes, err := json.Marshal(payload)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare request"})
        return
    }

    req, err := http.NewRequest("POST", tokenUrl, bytes.NewBuffer(payloadBytes))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
        return
    }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch access token"})
        return
    }
    defer resp.Body.Close()

    // Check response status
    if resp.StatusCode != http.StatusOK {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub API returned an error"})
        return
    }

    var tokenResponse map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse GitHub response"})
        return
    }

    accessToken, ok := tokenResponse["access_token"].(string)
    if !ok || accessToken == "" {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "No access token in GitHub response"})
        return
    }

    // Save the token securely for the user
    encryptedToken, err := lib.EncryptSecret(accessToken)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt access token"})
        return
    }
    user.Git.Token = encryptedToken
    models.DB.Save(&user)

	err= lib.ValidateGit(user)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Git credentials", "details": err.Error()})
		return
	}

    c.JSON(http.StatusOK, gin.H{"message": "GitHub linked successfully"})
}


func ValidateLink(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	AccessID, err := lib.DecryptSecret(user.AwsCreds.AccessKeyID)
	if(err != nil){
		log.Fatal(err)
	}
	AWSSecret, err := lib.DecryptSecret(user.AwsCreds.SecretAccessKey)
	if(err != nil){
		log.Fatal(err)
	}

	err = lib.ValidateAWSCredentials(AccessID, AWSSecret)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate AWS credentials", "details": err.Error()})
		return
	}
	err=lib.ValidateOpenAICredentials(user.OpenAICreds.APIKey)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate OAI credentials", "details": err.Error()})
		return
	}
	err=lib.ValidatePortfolioAPI(user.Portfolio.RootEndpoint,user.Portfolio.APIKey)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Portfolio credentials", "details": err.Error()})
		//return
	}

	encryptedPortfolioAPIKey, err := lib.EncryptSecret(user.Portfolio.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Portfolio API Key", "details": err.Error()})
		return
	}


	encryptedAccessKeyID, err := lib.EncryptSecret(user.AwsCreds.AccessKeyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt AWS Access Key ID", "details": err.Error()})
		return
	}

	encryptedSecretAccessKey, err := lib.EncryptSecret(user.AwsCreds.SecretAccessKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt AWS Secret Access Key", "details": err.Error()})
		return
	}

	encryptedApiKey, err := lib.EncryptSecret(user.OpenAICreds.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt OpenAI API Key", "details": err.Error()})
		return
	}

	encryptedPortfolioURL, err := lib.EncryptSecret(user.Portfolio.RootEndpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Portfolio Root Endpoint", "details": err.Error()})
		return
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
	c.JSON(http.StatusOK, gin.H{"message": "User Linked successfully"})
}
