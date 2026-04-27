package lib

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SpinAuthMiddleware middleware for Spin service
func SpinAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
            c.Abort()
            return
        }

        // Validate service token
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        valid, token, err := ValidateServiceToken(tokenString)
        if err != nil || !valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired service token"})
            c.Abort()
            return
        }

        // Extract project and user context
        projectID, err := token.GetString("project-id")
        userID, err := token.GetString("user-id")
        if projectID == "" || userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            c.Abort()
            return
        }

        // Set context for handler
        c.Set("projectID", projectID)
        c.Set("userID", userID)
        c.Next()
    }
}