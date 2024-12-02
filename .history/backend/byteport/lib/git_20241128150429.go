package lib

import (
	"byteport/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v32/github"
)
const (
	refreshInterval = 7 * time.Hour + 45 * time.Minute
	refreshChangeInterval = 3300* time.Hour
)
func ListRepositories(ctx *gin.Context, accessToken string) ([]*github.Repository, error) {
	client := github.NewClient(nil)
	client.Auth = &github.Auth{
		Token: accessToken,
		Type:  "bearer",
	}
	repos, _, err := client.Repositories.List(ctx, "", nil)
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

func GetUserAccessToken(apiURL, pasetoToken, code string) (models.Git, error) {
	valid, _, err := ValidateToken(pasetoToken)
	if err != nil || !valid {
		return nil, fmt.Errorf("failed to verify PASETO token: %v", err)
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
		return nil, fmt.Errorf("failed to get user access token: %v", err)
	}
	return auth, nil
}
func refreshToken(apiURL, pasetoToken, refreshToken string) (string, error) {
	var secrets models.GitSecret 
	models.DB.First(&secrets)
	url := fmt.Sprintf("%s/login/oauth/access_token?client_id=%s&client_secret=%s&grant_type=refresh_token&refresh_token=%s", apiURL, secrets.ClientID, secrets.ClientSecret, refreshToken)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+pasetoToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to GitHub API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to get Refresh Access Token: status code %d", resp.StatusCode)
	}

	// Parse the response to get the token
	var response struct {
		Token string `json:"access_token"`
		Expires_in int `json:"expires_in"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to parse Refresh Access token response: %v", err)
	}

	return response.Token, nil
}