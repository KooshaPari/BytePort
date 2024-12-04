package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/lib"
	"nvms/models"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)
func getArchiveURL(archiveURL string) string {
    // split by {archive_format} and {/ref}
    // replace {archive_format} with zip
    // replace {/ref} with /main
     splitURL := strings.Split(archiveURL, "/{")
    if len(splitURL) > 1 {
        return splitURL[0] + "/zip/main"
    }
    return archiveURL
}
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
        archiveURL := getArchiveURL(project.Repository.ArchiveURL)
        req, err := http.NewRequest("GET", archiveURL, nil)
        if err != nil {
            http.Error(w, "Failed to create request", http.StatusInternalServerError)
        }
        req.Header.Set("Authorization", "Bearer "+authToken)
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            http.Error(w, "Failed to get archive", http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        // Send response
        w.Header().Set("Content-Type", "application/json")
        // say success with repository worktree
        w.WriteHeader(http.StatusOK)
        io.Copy(w, resp.Body)
      
    })
}

func main() {}