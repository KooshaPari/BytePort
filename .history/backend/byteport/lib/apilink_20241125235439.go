package lib

import (
	"byteport/models"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

// ValidatePortfolioAPI validates the provided portfolio API key and endpoint.
func ValidatePortfolioAPI(rootEndpoint, apiKey string) error {
	fmt.Println("Validating Portfolio API...")
	req, err := http.NewRequest("GET", rootEndpoint+"/dev/templates", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Portfolio API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	fmt.Printf("Portfolio API Response: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid Portfolio API Key or URL. Status code: %d", resp.StatusCode)
	}

	if !strings.Contains(string(body), "expected_keyword_or_structure") {
		return fmt.Errorf("unexpected response from Portfolio API: %s", string(body))
	}

	fmt.Println("Portfolio API validated successfully.")
	return nil
}

// LinkWithGithub redirects the user to GitHub for app installation.
func LinkWithGithub(c *gin.Context, user models.User) {
	appName := "byteport-gh"
	redirectURL := fmt.Sprintf("https://github.com/apps/%s/installations/new?state=%s", appName, user.UUID)
	c.Redirect(http.StatusFound, redirectURL)
}

// ValidateGit validates the GitHub app installation and fetches repositories.
// ValidateGit validates the GitHub app installation and fetches repositories.
func ValidateGit(userID, installationID string) error {
	const apiURL = "https://api.github.com"

	// Fetch Git secrets from the database (global secrets for the GitHub App)
	var gitSecrets models.GitSecret
	result := models.DB.First(&gitSecrets)
	if result.Error != nil {
		return fmt.Errorf("failed to retrieve Git secrets from the database: %v", result.Error)
	}

	// Decrypt GitHub App secrets
	appID, err := DecryptSecret(gitSecrets.AppID)
	if err != nil {
		return fmt.Errorf("failed to decrypt App ID: %v", err)
	}
	privateKey, err := DecryptSecret(gitSecrets.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt private key: %v", err)
	}

	fmt.Println("Generating JWT for GitHub App authentication...")
	// Generate the GitHub App JWT
	jwtToken, err := GenerateJWT(appID, []byte(privateKey))
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %v", err)
	}

	fmt.Println("Fetching installation token...")
	// Exchange the JWT for an installation token (user-specific)
	installationToken, err := GetInstallationToken(apiURL, jwtToken, installationID)
	if err != nil {
		return fmt.Errorf("failed to get installation token: %v", err)
	}

	fmt.Println("Fetching repositories associated with the installation...")
	// Fetch repositories associated with the GitHub App installation
	reposURL := fmt.Sprintf("%s/installation/repositories", apiURL)
	req, err := http.NewRequest("GET", reposURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+installationToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to GitHub API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch repositories: status code %d", resp.StatusCode)
	}

	// Parse and log the response to list repositories
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Printf("GitHub Repositories Response: %s\n", string(body))
	fmt.Println("GitHub App connection validated successfully.")

	// Save the installation token to the user record for future authenticated requests
	var user models.User
	if err := models.DB.Where("uuid = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("failed to find user: %v", err)
	}

	encryptedToken, err := EncryptSecret(installationToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt installation token: %v", err)
	}

	user.Git.Token = encryptedToken
	user.Git.InstallID = installationID
	if err := models.DB.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to save GitHub installation data: %v", err)
	}

	return nil
}

// ValidateGitRepo validates a specific repository using the installation token.
func ValidateGitRepo(repoURL, installationToken string) error {
	fmt.Println("Validating GitHub repository...")

	cmd := exec.Command("git", "ls-remote", repoURL)
	cmd.Env = append(cmd.Env, "GIT_ASKPASS=echo "+installationToken)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	fmt.Printf("Git Command: git ls-remote %s\n", repoURL)
	fmt.Printf("Stdout: %s\n", stdout.String())
	fmt.Printf("Stderr: %s\n", stderr.String())

	if err != nil {
		return fmt.Errorf("failed to validate Git repository: %v. Stderr: %s", err, stderr.String())
	}

	fmt.Println("Git repository validated successfully.")
	return nil
}

// ValidateOpenAICredentials validates the OpenAI API credentials.
func ValidateOpenAICredentials(apiToken string) error {
	fmt.Println("Validating OpenAI credentials...")
	req, err := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to OpenAI API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read OpenAI response: %v", err)
	}
	fmt.Printf("OpenAI Response: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid OpenAI API Key. Status code: %d", resp.StatusCode)
	}

	fmt.Println("OpenAI credentials validated successfully.")
	return nil
}

// ValidateAWSCredentials validates AWS credentials.
func ValidateAWSCredentials(accessKey, secretKey string) error {
	fmt.Println("Validating AWS credentials...")
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Region:      aws.String("us-east-1"), // Default region for validation
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	svc := s3.New(sess)
	_, err = svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Printf("AWS Error: %s, Code: %s, Message: %s\n", awsErr.Error(), awsErr.Code(), awsErr.Message())
		}
		return fmt.Errorf("invalid AWS credentials: %v", err)
	}

	fmt.Println("AWS credentials validated successfully.")
	return nil
}