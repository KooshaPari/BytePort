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
func Authenticate(c *gin.Context){
	// extract token from cookie
	token, err := c.Cookie("authToken")
	if(err != nil){
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	// validate token and get user
	user, err := lib.AuthenticateRequest(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	c.Set("user", *user)

        c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"User": user,
		})
}
func LinkHandler(c *gin.Context) {
	// Retrieve the authenticated user object
	user := c.MustGet("user").(models.User)
	var req models.User

	// Bind JSON to the request model
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}
	err =lib.ValidateAWSCredentials(req.AwsCreds.AccessKeyID,req.AwsCreds.SecretAccessKey)
	err= lib.ValidateGitRepo(req.Git.RepoURL,req.Git.AuthMethod,req.Git.AuthKey,  req.Git.TargetDirectory)
	lib.ValidateOpenAICredentials(req.OpenAICreds.APIKey)
	lib.ValidatePortfolioAPI(req.Portfolio.RootEndpoint,req.Portfolio.APIKey)
	

	encryptedPortfolioAPIKey, err := lib.EncryptSecret(req.Portfolio.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Portfolio API Key", "details": err.Error()})
		return
	}

	encryptedRepoURL, err := lib.EncryptSecret(req.Git.RepoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Git Repo URL", "details": err.Error()})
		return
	}

	encryptedAuthMethod, err := lib.EncryptSecret(req.Git.AuthMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Git Auth Method", "details": err.Error()})
		return
	}

	encryptedAuthKey, err := lib.EncryptSecret(req.Git.AuthKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Git Auth Key", "details": err.Error()})
		return
	}

	encryptedTargetDirectory, err := lib.EncryptSecret(req.Git.TargetDirectory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Git Target Directory", "details": err.Error()})
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
		RepoURL:         encryptedRepoURL,
		AuthMethod:      encryptedAuthMethod,
		AuthKey:         encryptedAuthKey,
		TargetDirectory: encryptedTargetDirectory,
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

	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}
func Login(c *gin.Context){
	var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token."})
				return
			}
			c.SetCookie("authToken", token, 3600, "/", "localhost", false, true)
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
func Signup(c *gin.Context){
		var req models.SignupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token."})
			return
		}
		c.SetCookie("authToken", token, 3600, "/", "localhost", false, true)

		c.JSON(http.StatusCreated, newUser)
}