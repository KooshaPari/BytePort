package routes

import (
	"byteport/lib"
	"byteport/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from headers
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
            c.Abort()
            return
        }

        // Validate token and get user
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        valid, token, err := lib.ValidateToken(tokenString)
        if err != nil || !valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        userID,_ := token.GetString("user-id")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
            c.Abort()
            return
        }

        // Retrieve user from database
        var user models.User
        if err := models.DB.Where("uuid = ?", userID).First(&user).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
            c.Abort()
            return
        }

        // Set user in context
        c.Set("user", user)
        c.Next()
    }
}
func linkHandler(c *gin.Context) {
	// update user obj with new posted one (sans email pass name)
	user := c.MustGet("user").(models.User)
	var req models.LinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.AwsCreds = req.AwsCreds
	user.OpenAICreds = req.OpenAICreds
	user.Portfolio = req.Portfolio
	user.Git = req.Git
	models.DB.Save(&user)
	c.JSON(http.StatusOK, user)
	
}