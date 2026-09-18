package routes

import (
	"byteport/lib"
	"byteport/models"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Canonical LLM provider identifiers. Clients have shipped camelCase variants
// ("openAI") of both the provider name and the credential field; the backend
// stores and looks up the lower-case spellings.
const (
	defaultLLMProvider = "openai"
	defaultLLMModal    = "gpt-4o"
)

func RetrieveRepositories(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	decryptedToken, err := lib.DecryptSecret(user.Git.Token)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt Git token"})
		return
	}
	repoList, err := lib.ListRepositories(decryptedToken)
	if err != nil {
		fmt.Println("Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list repositories", "details": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", []byte(repoList))

}
func HandleCallback(c *gin.Context) {
	fmt.Println("Handling callback...")
	// printout full query string
	code := c.Query("code")
	state := c.Query("state")
	comps := strings.SplitN(state, "<BYTEPORT>", 2)
	if len(comps) < 2 || comps[0] == "" || comps[1] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}
	authToken := comps[0]
	userID := comps[1]

	// Validate state
	var user models.User
	if err := models.DB.Where("uuid = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid uuid parameter"})
		return
	}

	authDetails, err := lib.GetUserAccessToken(authToken, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user access token", "details": err.Error()})
		return
	}
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
	err = lib.ValidateGit(user)
	if err != nil {
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
	fmt.Println("USR: ", user)
	fmt.Println("C:", c)
	// Get the authenticated user for saving later
	authUser := c.MustGet("user").(models.User)

	// Resolve the LLM credential before validating anything. The lookup is
	// case-insensitive and AIProvider.UnmarshalJSON accepts the camelCase
	// `apiKey` spelling, because clients have sent both `openAI` and `apiKey`
	// while the canonical spellings are `openai` and `api_key`. Misses here
	// used to be handed to the OpenAI validator as an empty string, which
	// surfaced as "Failed to validate OAI credentials" on every setup attempt
	// and silently never persisted the key.
	providerName := strings.TrimSpace(user.LLMConfig.Provider)
	if providerName == "" {
		providerName = defaultLLMProvider
	}
	provider, ok := user.LLMConfig.ProviderKey(providerName)
	if !ok {
		// A drifted provider name must not be fatal: fall back to the
		// canonical provider before rejecting the request.
		provider, ok = user.LLMConfig.ProviderKey(defaultLLMProvider)
	}
	if !ok || strings.TrimSpace(provider.APIKey) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing OpenAI API key",
			"details": fmt.Sprintf("no provider entry carrying an api_key for %q in llmConfig.providers", providerName),
		})
		return
	}

	// Validate with unencrypted credentials
	err := lib.ValidateAWSCredentials(
		user.AwsCreds.AccessKeyID,
		user.AwsCreds.SecretAccessKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate AWS credentials", "details": err.Error()})
		return
	}

	err = lib.ValidateOpenAICredentials(provider.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate OAI credentials", "details": err.Error()})
		return
	}

	err = lib.ValidatePortfolioAPI(user.Portfolio.RootEndpoint, user.Portfolio.APIKey)
	if err != nil {
		fmt.Println("failed to validate portfolio")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Portfolio credentials", "details": err.Error()})
		return
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

	encryptedApiKey, err := lib.EncryptSecret(provider.APIKey)
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

	// Always persist under the canonical provider name, regardless of the
	// spelling the client used on the way in.
	modal := strings.TrimSpace(provider.Modal)
	if modal == "" {
		modal = defaultLLMModal
	}
	authUser.LLMConfig = models.LLM{
		Provider: defaultLLMProvider,
		Providers: map[string]models.AIProvider{
			defaultLLMProvider: {
				Modal:  modal,
				APIKey: encryptedApiKey,
			},
		}}

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
