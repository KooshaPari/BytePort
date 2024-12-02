package routes

import (
	"byteport/lib"
	"byteport/models"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)
func RetrieveRepositories(c *gin.Context) {
    user := c.MustGet("user").(models.User)
   
    decryptedToken, err := lib.DecryptSecret(user.Git.Token)
     fmt.Println("DECR: ",decryptedToken)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt Git token"})
        return
    }
    repoList, err := lib.ListRepositories(decryptedToken)
	if err != nil {
        fmt.Println("Error: ",err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list repositories", "details": err.Error()})
        return
	}
    

    c.Data(http.StatusOK, "application/json", []byte(repoList))

}
func HandleCallback(c *gin.Context) {
	fmt.Println("Handling callback...")
	// printout full query string
	//fmt.Println(c.Request.URL.RawQuery)
	code := c.Query("code")
    state := c.Query("state")
    comps := strings.Split(state, "<BYTEPORT>")
    authToken := comps[0]
    userID := comps[1]
    fmt.Println("Code: ", code)
    //fmt.Println("State: ", state)
    fmt.Println("AuthToken: ", authToken)
    fmt.Println("UserID: ", userID)
    // Validate state
    var user models.User
    if err := models.DB.Where("uuid = ?", userID).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid uuid parameter"})
        return
    }
    /*valid, _, err := lib.ValidateToken(authToken)
    if err != nil || !valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }
    
	
    // Exchange code for user access tok
    fmt.Println("code: ", code)
    authDetails, err := lib.GetUserAccessToken(authToken, code)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user access token", "details": err.Error()})
    }
    fmt.Println("AuthDetails: ", authDetails)
    encryptedToken, err := lib.EncryptSecret(authDetails.Token)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt access token"})
        return
    }
    encryptedRefreshToken, err := lib.EncryptSecret(authDetails.RefreshToken)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt refresh token"})
        return
    }
   
   
    user.Git = models.Git{
        Token:              encryptedToken,
        RefreshToken:       encryptedRefreshToken,
        TokenExpiry:        authDetails.TokenExpiry,
        RefreshTokenExpiry: authDetails.RefreshTokenExpiry,
    }
    models.DB.Save(&user)
    fmt.Println("User access token saved successfully.")
	err= lib.ValidateGit(user)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Git credentials", "details": err.Error()})
		return
	}

    c.Header("Content-Type", "text/html")
    c.String(http.StatusOK, `
        <html>
        <body>
            <script>
                window.opener.postMessage('github-linked', '*');
                window.close();
            </script>
            <p>GitHub linked successfully. This window will close automatically.</p>
        </body>
        </html>
    `)
}


func ValidateLink(c *gin.Context) {
	var user models.User
    if err := c.BindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // Get the authenticated user for saving later
    authUser := c.MustGet("user").(models.User)

    // Validate with unencrypted credentials
    err := lib.ValidateAWSCredentials(
        user.AwsCreds.AccessKeyID, 
        user.AwsCreds.SecretAccessKey,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate AWS credentials", "details": err.Error()})
        return
    }

    err = lib.ValidateOpenAICredentials(user.OpenAICreds.APIKey)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate OAI credentials", "details": err.Error()})
        return
    }

    err = lib.ValidatePortfolioAPI(user.Portfolio.RootEndpoint, user.Portfolio.APIKey)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Portfolio credentials", "details": err.Error()})
        //return
    }

    // Encrypt credentials after validation
    encryptedAccessKeyID, err := lib.EncryptSecret(user.AwsCreds.AccessKeyID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt AWS Access Key ID"})
        return
    }

    encryptedSecretAccessKey, err := lib.EncryptSecret(user.AwsCreds.SecretAccessKey)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt AWS Secret Access Key"})
        return
    }

    encryptedApiKey, err := lib.EncryptSecret(user.OpenAICreds.APIKey)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt OpenAI API Key"})
        return
    }

    encryptedPortfolioURL, err := lib.EncryptSecret(user.Portfolio.RootEndpoint)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Portfolio Root Endpoint"})
        return
    }

    encryptedPortfolioAPIKey, err := lib.EncryptSecret(user.Portfolio.APIKey)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Portfolio API Key"})
        return
    }

    // Update the auth user with encrypted credentials
    authUser.AwsCreds = models.AwsCreds{
        AccessKeyID:     encryptedAccessKeyID,
        SecretAccessKey: encryptedSecretAccessKey,
    }

    authUser.OpenAICreds = models.OpenAICreds{
        APIKey: encryptedApiKey,
    }

    authUser.Portfolio = models.Portfolio{
        RootEndpoint: encryptedPortfolioURL,
        APIKey:       encryptedPortfolioAPIKey,
    }

    // Save the updated user
    if err := models.DB.Save(&authUser).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Credentials validated and saved successfully"})
}

  /*clientID,err := lib.DecryptSecret(secrets.ClientID)
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
    }*/