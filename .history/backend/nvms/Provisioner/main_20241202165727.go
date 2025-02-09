package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/lib"
	"nvms/models"
	"os"
	"path/filepath"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func downloadRepository(url string, token string, destPath string) error {
    // Create HTTP client with auth
    client := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return err
    }
    
    req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
    
    // Make request
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Create destination directory
    err = os.MkdirAll(destPath, 0755)
    if err != nil {
        return err
    }

    // Write zip file
    zipPath := filepath.Join(destPath, "repo.zip")
    f, err := os.Create(zipPath)
    if err != nil {
        return err
    }
    defer f.Close()

    _, err = io.Copy(f, resp.Body)
    return err
}

func init() {
    spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received request")
        var project models.Project
        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Error reading request body", http.StatusInternalServerError)
            return
        }
        defer r.Body.Close()

        err = json.Unmarshal(body, &project)
        if err != nil {
            http.Error(w, "Error parsing JSON", http.StatusBadRequest)
            return
        }

        // Get archive URL instead of git clone URL
        archiveURL := fmt.Sprintf("%s/archive/refs/heads/main.zip", 
            strings.TrimSuffix(project.Repository.MirrorURL, ".git"))

        userEncryptedToken := project.User.Git.Token
        authToken, err := lib.DecryptSecret(userEncryptedToken)
        if err != nil {
            http.Error(w, "Failed to decrypt user access token", http.StatusInternalServerError)
            return
        }

        destPath := fmt.Sprintf("/tmp/%s", project.UUID)
        err = downloadRepository(archiveURL, authToken, destPath)
        if err != nil {
            http.Error(w, "Failed to download repository err: "+err.Error(), http.StatusInternalServerError)
            return
        }

        fmt.Printf("Downloaded repository to: %s\n", destPath)
        w.WriteHeader(http.StatusOK)
    })
}

func main() {}