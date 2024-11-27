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

	
	fmt.Println("Redirecting user to GitHub App installation...")
	ValidateLink(c)
	lib.LinkWithGithub(c, user)
	lib.ValidateGit(user)
	err= lib.ValidateGit(user)
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate Git credentials", "details": err.Error()})
		return
	}
	encryptedGitToken, err := lib.EncryptSecret(req.Git.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt Git Token", "details": err.Error()})
		return
	}
	// Update user fields
	user.Git = models.Git{
		Token: encryptedGitToken,
	}

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