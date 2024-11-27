package routes

import (
	"byteport/lib"
	"byteport/models"
	"fmt"
	"log"
	"net/http"
	"os"

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
	// update user obj with new posted one (sans email pass name)
	user := c.MustGet("user").(models.User)
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(req)
	encrypted portfolioAPIKey, _ = lib.EncryptSecret(req.Portfolio.APIKey, os.Getenv("ENCRYPTION_KEY"))
	encryptedRepoURL, _ := lib.EncryptSecret(req.Git.RepoURL, os.Getenv("ENCRYPTION_KEY"))
	encryptedAuthMethod, _ := lib.EncryptSecret(req.Git.AuthMethod, os.Getenv("ENCRYPTION_KEY"))
	encryptedAuthKey, _ := lib.EncryptSecret(req.Git.AuthKey, os.Getenv("ENCRYPTION_KEY"))
	encryptedTargetDirectory, _ := lib.EncryptSecret(req.Git.TargetDirectory, os.Getenv("ENCRYPTION_KEY"))

	user.Git = models.Git{
		RepoURL:         encryptedRepoURL,
		AuthMethod:      encryptedAuthMethod,
		AuthKey:         encryptedAuthKey,
		TargetDirectory: encryptedTargetDirectory,
	}
	encryptedAccessKeyID, _ := lib.EncryptSecret(req.AwsCreds.AccessKeyID, os.Getenv("ENCRYPTION_KEY"))
	encryptedSecretAccessKey, _ := lib.EncryptSecret(req.AwsCreds.SecretAccessKey, os.Getenv("ENCRYPTION_KEY"))
	encryptedSessionToken, _ := lib.EncryptSecret(req.AwsCreds.SessionToken, os.Getenv("ENCRYPTION_KEY"))

	user.AwsCreds = models.AwsCreds{
		AccessKeyID:     encryptedAccessKeyID,
		SecretAccessKey: encryptedSecretAccessKey,
	}
	encryptedApiKey, _ := lib.EncryptSecret(req.OpenAICreds.ApiKey, os.Getenv("ENCRYPTION_KEY"))

	user.OpenAICreds = models.OpenAICreds{
		APIKey: encryptedApiKey,
	}
	encryptedPortfolioURL, _ := lib.EncryptSecret(req.Portfolio.PortfolioURL, os.Getenv("ENCRYPTION_KEY"))

	user.Portfolio = models.Portfolio{
		RootEndpoint: encryptedPortfolioURL,
		APIKey: encryptedPor
	}
	c.JSON(http.StatusOK, user)
	
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