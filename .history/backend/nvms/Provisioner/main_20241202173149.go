package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/lib"
	"nvms/models"

	"github.com/go-git/go-billy/v5/memfs"


	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
    spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Received request")
        
        // Parse request body
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

        // Decrypt token
        authToken, err := lib.DecryptSecret(project.User.Git.Token)
        if err != nil {
            http.Error(w, "Failed to decrypt user access token", http.StatusInternalServerError)
            return
        }

        // Setup in-memory filesystem
        fs := memfs.New()
        storage := memory.NewStorage()

        // Clone repository
        repo, err := git.Clone(storage, fs, &git.CloneOptions{
            URL: project.Repository.CloneURL,
            Auth: &githttp.BasicAuth{
                Username: "git",  // Using "git" instead of "blank" as some Git servers prefer this
                Password: authToken,
            },
            Progress: nil, // Disable progress as it might cause issues with WASI
            SingleBranch: true,
            Depth: 1, // Shallow clone for efficiency
        })
        if err != nil {
            fmt.Printf("Clone error: %v\n", err)
            http.Error(w, fmt.Sprintf("Failed to clone repository: %v", err), http.StatusInternalServerError)
            return
        }

        // Get the worktree
        w_tree, err := repo.Worktree()
        if err != nil {
            http.Error(w, "Failed to get worktree", http.StatusInternalServerError)
            return
        }

        // List files in root directory
        files, err := w_tree.Filesystem.ReadDir("/")
        if err != nil {
            http.Error(w, "Failed to read repository contents", http.StatusInternalServerError)
            return
        }

        // Create response with file list
        var fileList []string
        for _, file := range files {
            fileList = append(fileList, file.Name())
        }

        response := map[string]interface{}{
            "status": "success",
            "files":  fileList,
        }

        // Send response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    })
}

func main() {}