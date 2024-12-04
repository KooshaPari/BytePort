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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func downloadRepository(url string, token string, destPath string) error {
    // Create HTTP client with auth
    client := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return err
    }
    
    req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
    
    // Make request
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Failed to make request:", err)
        return err
    }
    defer resp.Body.Close()

    // Create destination directory
    err = os.MkdirAll(destPath, 0755)
    if err != nil {
        fmt.Println("Failed to create destination directory:", err)
        return err
    }

    // Write zip file
    zipPath := filepath.Join(destPath, "repo.zip")
    f, err := os.Create(zipPath)
    if err != nil {
        fmt.Println("Failed to create zip file:", err)
        return err
    }
    defer f.Close()

    _, err = io.Copy(f, resp.Body)
    if err != nil {
        fmt.Println("Failed to write zip file:", err)
        return err
    }
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

        archiveURL := project.Repository.CloneURL
        
        res, err := git.PlainClone(memory.NewStorage(), false, &git.CloneOptions{
            URL: archiveURL,
            // User Acess token in Authorization bearer
            Auth: &http.BasicAuth{
                Username: "blank",
                Password: project.User.Git.Token,
            },
            Progress: os.Stdout,
        })
        if err != nil {
            http.Error(w, "Failed to clone repository", http.StatusInternalServerError)
            return
        }
       

        

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