package routes

import (
	"byteport/lib"
	"byteport/models"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// currentUser resolves the authenticated user that AuthMiddleware stored in
// the gin context. When the context carries no usable user it writes the error
// response itself and reports false, so protected handlers cannot forget to
// check the session before trusting request data.
func currentUser(c *gin.Context) (models.User, bool) {
	value, exists := c.Get("user")
	if !exists {
		respondUnauthorized(c, "Unauthorized")
		return models.User{}, false
	}
	user, ok := value.(models.User)
	if !ok {
		respondInternalError(c, "Invalid user context")
		return models.User{}, false
	}
	return user, true
}

func setAuthCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("authToken", token, 3600, "/", c.GetHeader("Host"), true, true)
}

func Authenticate(c *gin.Context) {
	// extract token from cookie

	token, err := c.Cookie("authToken")
	if err != nil {
		respondUnauthorized(c, "Unauthorized")
		return
	}

	// validate token and get user
	user, err := lib.AuthenticateRequest(token)
	if err != nil {
		respondUnauthorized(c, "Unauthorized")
		return
	}
	c.Set("user", *user)

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"User":    user,
	})
}
func LinkHandler(c *gin.Context) {
	// Retrieve the authenticated user object
	user, ok := currentUser(c)
	if !ok {
		return
	}
	fmt.Println("Linking with Github: ", user)

	lib.LinkWithGithub(c, user)
	fmt.Println("Validating Details")

}
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	var user models.User
	if err := models.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Failed, user not found"})
		return
	}

	if lib.ValidatePass(req.Password, user.Password) {
		// set cookie
		token, err := lib.GenerateToken(user)
		if err != nil {
			log.Printf("Error generating token: %v", err)
			respondInternalError(c, "Failed to generate authentication token.")
			return
		}
		setAuthCookie(c, token)
		user.Password = ""
		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"user":    user,
		})
	} else {
		fmt.Println("Invalid Credentials.")
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Failed, invalid credentials",
		})
	}
}
func Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	hash := lib.EncryptPass(req.Password)

	// Check for pre-existing user
	var existingUser models.User
	if err := models.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		// User already exists
		c.JSON(http.StatusConflict, gin.H{
			"message": "Failed, User Already Exists",
		})
		return
	}

	newUser := models.User{
		UUID:     uuid.NewString(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hash),
	}

	// Create the new user
	if err := models.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create user",
			"error":   err.Error(),
		})
		return
	}

	newUser.Password = ""
	token, err := lib.GenerateToken(newUser)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		respondInternalError(c, "Failed to generate authentication token.")
		return
	}
	setAuthCookie(c, token)

	c.JSON(http.StatusCreated, newUser)
}

func UpdateLink(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	if user.LLMConfig.Provider == "" {
		user.LLMConfig = models.LLM{
			Provider:  defaultLLMProvider,
			Providers: make(map[string]models.AIProvider),
		}
	}
	if user.LLMConfig.Providers == nil {
		user.LLMConfig.Providers = make(map[string]models.AIProvider)
	}
	decryptedOAI := "err"
	var err error
	// ProviderEntry matches case-insensitively so a credential stored under a
	// drifted key ("openAI") is still found, and it reports the key it came
	// from so the entry is written back in place instead of duplicated.
	providerKey, provider, providerFound := user.LLMConfig.ProviderEntry(user.LLMConfig.Provider)
	if user.LLMConfig.Provider != "local" {
		if !providerFound {
			respondErrorWithDetails(c, http.StatusInternalServerError, "Failed to decrypt OAI", gin.H{
				"details": fmt.Sprintf("no provider entry for %q", user.LLMConfig.Provider),
			})
			return
		}
		decryptedOAI, err = lib.DecryptSecret(provider.APIKey)
		if err != nil {
			respondInternalError(c, "Failed to decrypt OAI")
			return
		}
	}
	decryptedAWSAccess, err := lib.DecryptSecret(user.AwsCreds.AccessKeyID)
	if err != nil {
		respondInternalError(c, "Failed to decrypt AWS Access")
		return
	}
	decryptedAWSSecret, err := lib.DecryptSecret(user.AwsCreds.SecretAccessKey)
	if err != nil {
		respondInternalError(c, "Failed to decrypt AWS Secret")
		return
	}
	decryptedPortfolioURL, err := lib.DecryptSecret(user.Portfolio.RootEndpoint)
	if err != nil {
		respondInternalError(c, "Failed to decrypt Portfolio URL")
		return
	}
	decryptedPortfolioKey, err := lib.DecryptSecret(user.Portfolio.APIKey)
	if err != nil {
		respondInternalError(c, "Failed to decrypt Portfolio Key ")
		return
	}
	user.AwsCreds = models.AwsCreds{
		AccessKeyID:     decryptedAWSAccess,
		SecretAccessKey: decryptedAWSSecret,
	}
	user.Portfolio = models.Portfolio{
		RootEndpoint: decryptedPortfolioURL,
		APIKey:       decryptedPortfolioKey,
	}
	if !providerFound {
		// Provider was "local" (no credential entry) or absent: keep the key
		// the request used so nothing is invented.
		providerKey = user.LLMConfig.Provider
	}
	provider.APIKey = decryptedOAI
	user.LLMConfig.Providers[providerKey] = provider
	// return new user obj
	user.Password = ""
	c.JSON(http.StatusOK, user)

}
func UpdateUser(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Password != "" {
		hash := lib.EncryptPass(req.Password)
		user.Password = string(hash)
	}
	if err := models.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update user",
			"error":   err.Error(),
		})
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, user)

}
