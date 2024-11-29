package lib

import (
	"byteport/models"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v67/github"
)
const (
	refreshInterval = 7 * time.Hour + 45 * time.Minute
	refreshChangeInterval = 3300* time.Hour
)
func ListRepositories(ctx *gin.Context, accessToken string) ([]*github.Repository, error) {
	
	client := github.NewClient(nil).WithAuthToken(accessToken)
	var user models.User = c.MustGet("user").(models.User)
	repos, _, err := client.Repositories.ListByUser(ctx,client.Users., nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %v", err)
	}
	return repos, nil
}
// LinkWithGithub redirects the user to GitHub for app installation.
func LinkWithGithub(c *gin.Context, user models.User) {
	//appName := "byteport-gh"
	var secrets models.GitSecret 
	models.DB.First(&secrets)
	authToken, err := GenerateGitPaseto(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PASETO token"})
		return
	}
	ClientID, err := DecryptSecret(secrets.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt client id"})
		return
	}
	state := authToken + "<BYTEPORT>" + user.UUID
	redirectURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&state=%s", ClientID, state)
	fmt.Println("Redirecting user to GitHub App installation..." + redirectURL)
	c.Redirect(http.StatusFound, redirectURL)
	
}

func GenerateGitPaseto(user models.User) (string, error) {

	token, err := GenerateToken(user)
    if err != nil {
        return "", fmt.Errorf("failed to generate PASETO token: %v", err)
    }

    return token, nil
}

func GetUserAccessToken(ctx *gin.Context, pasetoToken, code string)(*github.Authorization, error) {
	valid, _, err := ValidateToken(pasetoToken)
	if err != nil || !valid {
		return models.Git{}, fmt.Errorf("failed to verify PASETO token: %v", err)
	}
	secrets := models.GitSecret{}
	models.DB.First(&secrets)
	client := github.NewClient(nil)
	client.Auth = &github.Auth{
		Token: pasetoToken,
		Type:  "bearer",
	}
	auth, _, err := client.Authorizations.Create(ctx, &github.AuthorizationRequest{
		ClientID: secrets.ClientID,
		Code:     code,
	})
	if err != nil {
		return models.Git{}, fmt.Errorf("failed to get user access token: %v", err)
	}
	return auth, nil
}
func RefreshToken(ctx *gin.Context, pasetoToken, refreshToken string) (*github.Authorization, error) {
	secrets := models.GitSecret{}
	models.DB.First(&secrets)
	client := github.NewClient(nil)
	client.Auth = &github.Auth{
		Token: pasetoToken,
		Type:  "bearer",
	}
	auth, _, err := client.Authorizations.Update(ctx, &github.AuthorizationRequest{
		ClientID: secrets.ClientID,
		RefreshToken: refreshToken,
		GrantType: "refresh_token",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}
	return auth, nil
}