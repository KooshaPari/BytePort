package routes

import (
	"archive/zip"
	"byteport/lib"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// Main function to download and extract the GitHub repository
func TestZip(c *gin.Context) {
	archiveURL := "https://api.github.com/repos/KooshaPari/BytePort/zipball/main"
	var Project models.Project
	err := c.ShouldBindJSON(&Project)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User

	encryptedToken := user.Git.Token
	authToken, err := lib.DecryptSecret(encryptedToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt Git token"})
		return
	}

	// Download the repository archive
	fmt.Println("Downloading repository archive...")
	zipResp, err := downloadRepo(archiveURL, authToken)
	if err != nil {
		fmt.Printf("Error downloading archive: %v\n", err)
		return
	}
	defer zipResp.Body.Close()

	// Process the ZIP file
	fmt.Println("Processing ZIP archive...")
	zipReader, err := processZip(zipResp)
	if err != nil {
		fmt.Printf("Error processing ZIP archive: %v\n", err)
		return
	}

	// Extract and read the files
	fileList, fileMap := readZip(zipReader)
	fmt.Println("Extracted files:")
	fmt.Println("Filemap: ", fileMap["README.md"])
	for _, file := range fileList {
		fmt.Println(" -", file)
	}
	fmt.Printf("Total files: %d\n", len(fileList))
}

// Function to download the repository archive
func downloadRepo(archiveURL, authToken string) (*http.Response, error) {
	req, err := http.NewRequest("GET", archiveURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("User-Agent", "Byteport")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	req.Header.Add("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch archive: %w", err)
	}

	// Handle redirects if needed
	if resp.StatusCode == http.StatusFound {
		redirectURL := resp.Header.Get("Location")
		if redirectURL == "" {
			return nil, fmt.Errorf("no redirect URL provided")
		}
		resp.Body.Close() // Close the previous response body
		req, err = http.NewRequest("GET", redirectURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create redirect request: %w", err)
		}
		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to follow redirect: %w", err)
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}
	return resp, nil
}

// Function to process the ZIP file in memory
func processZip(zipResp *http.Response) (*zip.Reader, error) {
	zipBytes, err := io.ReadAll(zipResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read zip file: %w", err)
	}
	if len(zipBytes) == 0 {
		return nil, fmt.Errorf("received empty zip file")
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to read zip archive: %w", err)
	}
	return zipReader, nil
}

// Function to extract and read the ZIP file
func readZip(zipReader *zip.Reader) ([]string, map[string][]byte) {
	fileMap := make(map[string][]byte)
	var fileList []string
	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		parts := strings.Split(file.Name, "/")
		if len(parts) > 1 {
			parts = parts[1:]
		}
		relativePath := path.Join(parts...)
		fmt.Printf("File: %s, Method: %d\n", file.Name, file.Method)

		if file.Method != zip.Deflate && file.Method != zip.Store {
			fmt.Printf("Warning: Unsupported compression method %d for file %s\n", file.Method, file.Name)
			continue
		}
		rc, err := file.Open()
		if err != nil {
			fmt.Printf("Warning: Could not open file %s: %v\n", file.Name, err)
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			fmt.Printf("Warning: Could not read file %s: %v\n", file.Name, err)
			continue
		}

		fileMap[relativePath] = content
		fileList = append(fileList, relativePath)
	}
	return fileList, fileMap
}