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

	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"gopkg.in/src-d/go-billy.v4/memfs"

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
         userEncryptedToken := project.User.Git.Token
        authToken, err := lib.DecryptSecret(userEncryptedToken)
        if err != nil {
            http.Error(w, "Failed to decrypt user access token", http.StatusInternalServerError)
            return
        }
        dest := filepath.Join("tmp", project.Name, "-", project.UUID)
        destFS := memfs.New()
        res, err := git.Clone(memory.NewStorage(), dest, &git.CloneOptions{

            // User Acess token in Authorization bearer
            Auth: &githttp.BasicAuth{
                Username: "blank",
                Password: authToken,
            },
            URL: archiveURL,
            Progress: os.Stdout,
        })
        if err != nil {
            http.Error(w, "Failed to clone repository", http.StatusInternalServerError)
            return
        }
        fmt.Println(res.Log(&git.LogOptions{
            All: true,}))
       

        fmt.Printf("Downloaded repository to\n")
        w.WriteHeader(http.StatusOK)
    })
}

func main() {}