package lib

import (
	"byteport/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)
const (
	refreshInterval = 7 * time.Hour + 45 * time.Minute
	refreshChangeInterval = 3300* time.Hour
)
func ListRepositories(accessToken string) (string, error) {
	const apiURL = "https://api.github.com"
	
	url := fmt.Sprintf("%s/user/repos", apiURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+ accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	fmt.Println("HEADER: ", req.Header)
	fmt.Println("Acess token: ", accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to GitHub API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch repositories: status code %d, body: %s", resp.StatusCode, body)
	}

	return string(body), nil
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
	if(err != nil){
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt client id"})
		return
	}
	// state has paseto token and user id
	var state string = authToken + "<BYTEPORT>"+ user.UUID
	redirectURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&state=%s", ClientID, state)
	fmt.Println("Redirecting user to GitHub App installation..."+ redirectURL )
	c.Redirect(http.StatusFound, redirectURL)
	
}

func GenerateGitPaseto(user models.User) (string, error) {

	token, err := GenerateToken(user)
    if err != nil {
        return "", fmt.Errorf("failed to generate PASETO token: %v", err)
    }

    return token, nil
}

func GetUserAccessToken( pasetoToken, code string) (models.Git, error) {
	const apiURL = "https://github.com/login/oauth/access_token"
	 valid, _, err := ValidateToken(pasetoToken)
    if err != nil || !valid {
        return models.Git{}, fmt.Errorf("failed to verify PASETO token: %v", err)
    }
	var secrets models.GitSecret 
	models.DB.First(&secrets)
	clientID,err := DecryptSecret(secrets.ClientID)
	if err != nil {
		return models.Git{}, fmt.Errorf("failed to decrypt client id: %v", err)
	
	}
    clientSecret,err := DecryptSecret(secrets.ClientSecret)
	if err != nil {
		return models.Git{}, fmt.Errorf("failed to decrypt client secret: %v", err)
	
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
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return models.Git{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
    req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch access token"})
        return models.Git{}, fmt.Errorf("failed to connect to GitHub API: %v", err)
		
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


	// Parse the response to get the token
	var response models.Git = models.Git{
		Token: "", 
		RefreshToken: "",
		TokenExpiry: time.Now(),
		RefreshTokenExpiry: time.Now(),
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return models.Git{}, fmt.Errorf("failed to parse User Access token response: %v", err)
	}else{
		fmt.Println(resp.Body)
		fmt.Println("response token ", response.Token)
		fmt.Println("response refresh token ", response.RefreshToken)
		fmt.Println("response token expiry ", response.TokenExpiry)
		fmt.Println("response refresh token expiry ", response.RefreshTokenExpiry)
	}
	response.TokenExpiry = time.Now().Add(7*time.Hour + 45*time.Minute) // Set access token expiry
	response.RefreshTokenExpiry = time.Now().Add(4*30*24*time.Hour + 12*time.Hour) // Set refresh token expiry (4.5 months)

	return response, nil
}
func refreshToken(user models.User, pasetoToken string) (string, error) {
	const apiURL = "https://api.github.com"
	var secrets models.GitSecret 
	decryptedToken, err := DecryptSecret(user.Git.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt refresh token: %v", err)
	}
	models.DB.First(&secrets)
	
			// Logic to refresh the access token
			// Call your function to refresh the token here
			// Update the user record with the new token and expiry
			url := fmt.Sprintf("%s/login/oauth/access_token?client_id=%s&client_secret=%s&grant_type=refresh_token&refresh_token=%s", apiURL, secrets.ClientID, secrets.ClientSecret, decryptedToken)
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
		
	
	// GitHub API to generate installation token
	return "nil", nil
	
}
func refreshTokens(){
			var users []models.User
	models.DB.Find(&users)

	for _, user := range users {
		if time.Now().After(user.Git.TokenExpiry) {
			// Logic to refresh the access token
			// Call your function to refresh the token here
			// Update the user record with the new token and expiry
			token, err := GenerateGitPaseto(user)
			if err != nil {
				fmt.Println(err)
			}
			newToken, err := refreshToken(user,token)
			if err != nil {
				fmt.Println(err)
			}
			encryptedToken, err := EncryptSecret(newToken)
			if err != nil {
				fmt.Println(err)
			}
			// Update the user record with the new token and expiry
			models.DB.Model(&user).Update("git_token", encryptedToken)
			
		}
	}
}
func StartTokenRefreshJob() {
	ticker := time.NewTicker(7*time.Hour + 45*time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			refreshTokens()

		}
	}
}
