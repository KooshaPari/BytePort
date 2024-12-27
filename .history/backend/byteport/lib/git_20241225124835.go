package lib

import (
	"byteport/models"
	"bytes"
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
       return models.Git{}, fmt.Errorf("failed to marshal payload: %v", err)
    }
	 req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
    if err != nil {
       return models.Git{}, fmt.Errorf("failed to create request: %v", err)
    }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {

        return models.Git{}, fmt.Errorf("failed to connect to GitHub API: %v", err)

    }
    defer resp.Body.Close()

    // Check response status
    if resp.StatusCode != http.StatusOK {
       return models.Git{}, fmt.Errorf("failed to get Access Token: status code %d", resp.StatusCode)
    }

    var tokenResponse map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
        return models.Git{}, fmt.Errorf("failed to parse GitHub response: %v", err)
    }


	// Parse the response to get the token
	var response models.Git = models.Git{
		Token: "", 
		RefreshToken: "",
		TokenExpiry: time.Now(),
		RefreshTokenExpiry: time.Now(),
	}
	response.Token = tokenResponse["access_token"].(string)
	response.RefreshToken = tokenResponse["refresh_token"].(string)
	
	response.TokenExpiry = time.Now().Add(7*time.Hour + 45*time.Minute) // Set access token expiry
	response.RefreshTokenExpiry = time.Now().Add(4*30*24*time.Hour + 12*time.Hour) // Set refresh token expiry (4.5 months)

	return response, nil
}
func refreshToken(user models.User, pasetoToken string) (models.Git, error) {
	const apiURL = "https://github.com/login/oauth/access_token"
	var secrets models.GitSecret 
	fmt.Println("Refreshing Token - Decrypt")
	decryptedToken, err := DecryptSecret(user.Git.RefreshToken)
	if err != nil {
		return  models.Git{}, fmt.Errorf("failed to decrypt refresh token: %v", err)
	}
	models.DB.First(&secrets)
	clientID,err := DecryptSecret(secrets.ClientID)
	if err != nil {
		return models.Git{}, fmt.Errorf("failed to decrypt client id: %v", err)
	
	}
    clientSecret,err := DecryptSecret(secrets.ClientSecret)
	if err != nil {
		return  models.Git{}, fmt.Errorf("failed to decrypt client secret: %v", err)
	
	}
	fmt.Println("Refreshing Token - Req")
	payload := map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"grant_type": "refresh_token",
			"refresh_token":          decryptedToken,
		}
      payloadBytes, err := json.Marshal(payload)
    if err != nil {
       return  models.Git{}, fmt.Errorf("failed to marshal payload: %v", err)
    }
	 req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
    if err != nil {
       return  models.Git{}, fmt.Errorf("failed to create request: %v", err)
    }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	fmt.Println("Refreshing Token - Send")
    resp, err := client.Do(req)
    if err != nil {

        return models.Git{}, fmt.Errorf("failed to connect to GitHub API: %v", err)

    }
    defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return  models.Git{}, fmt.Errorf("failed to get Refresh Access Token: status code %d", resp.StatusCode)
	}


	var tokenResponse map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
        return models.Git{}, fmt.Errorf("failed to parse GitHub response: %v", err)
    }


	// Parse the response to get the token
	var response models.Git = models.Git{
		Token: "", 
		RefreshToken: "",
		TokenExpiry: time.Now(),
		RefreshTokenExpiry: time.Now(),
	}
	fmt.Println("Refreshing Token - Parse")
	fmt.Println("Response Token: ", response.Token)
	fmt.Println("Response Refresh Token: ", response.RefreshToken)
	fmt.Println("Response Expiry: ", response.RefreshTokenExpiry)
	fmt.Println("Response Refresh Expiry: ", response.RefreshTokenExpiry)
	response.Token = tokenResponse["access_token"].(string)
	response.RefreshToken = tokenResponse["refresh_token"].(string)
	
	response.TokenExpiry = time.Now().Add(7*time.Hour + 45*time.Minute) // Set access token expiry
	response.RefreshTokenExpiry = time.Now().Add(4*30*24*time.Hour + 12*time.Hour) // Set refresh token expiry (4.5 months)

	return response, nil
}
func refreshTokens(){
			var users []models.User
	models.DB.Find(&users)

	for _, user := range users {
		if time.Now().After(user.Git.TokenExpiry) {
			fmt.Println("Refreshing access token for user: ", user.UUID)
			// Logic to refresh the access token
			// Call your function to refresh the token here
			// Update the user record with the new token and expiry
			token, err := GenerateGitPaseto(user)
			if err != nil {
				fmt.Println("Err getting Paseto: ",err)
			}
			fmt.Println("Got Token")
			newToken, err := refreshToken(user,token)
			if err != nil {
				fmt.Println("Err Refreshing:",err)
			}
			fmt.Println("New Token Genned")
			encryptedToken, err := EncryptSecret(newToken.Token)
			if err != nil {
				fmt.Println(err)
			}
			encryptedRefreshToken, err := EncryptSecret(newToken.RefreshToken)
			if err != nil {
				fmt.Println(err)
			}
			user.Git = models.Git{
				Token: encryptedToken,
				RefreshToken: encryptedRefreshToken,
				TokenExpiry: newToken.TokenExpiry,
				RefreshTokenExpiry: newToken.RefreshTokenExpiry,
			}
			// Update the user record with the new token and expiry
			models.DB.Save(&user)
			
		}
	}
}
func StartTokenRefreshJob() {
	refreshTokens()
	ticker := time.NewTicker(7*time.Hour + 45*time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("Checking For Git Token Refresh")
		refreshTokens()
	}
}
