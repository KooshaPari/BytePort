package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/lib"
	"nvms/models"
	"path"
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)
func getArchiveURL(archiveURL string) string {
    // split by {archive_format} and {/ref}
    // replace {archive_format} with zip
    // replace {/ref} with /main
    fmt.Println("Archive URL: ", archiveURL)
     splitURL := strings.Split(archiveURL, "/{")
    if len(splitURL) > 1 {
        return splitURL[0] + "/zipball/main"
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
        fmt.Println("Archive URL: ", archiveURL)
        req, err := http.NewRequest("GET", archiveURL, nil)
        if err != nil {
            http.Error(w, "Failed to create request", http.StatusInternalServerError)
        }
        req.Header.Set("Authorization", "Bearer "+authToken)
        req.Header.Set("Accept", "application/vnd.github+json")
        client := &http.Client{
            CheckRedirect: nilm
        }
        resp, err := client.Do(req)
        if err != nil {
            http.Error(w, "Failed to get archive", http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()
        // save locally
         // Read the entire zip file into memory
        zipBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            http.Error(w, "Failed to read zip file", http.StatusInternalServerError)
            return
        }

        // Create a reader for the zip file
        zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
        if err != nil {
            http.Error(w, "Failed to read zip archive", http.StatusInternalServerError)
            return
        }

        // Process each file in the zip
        fileMap := make(map[string][]byte)
        var fileList []string
        
        for _, file := range zipReader.File {
            // Skip directories
            if file.FileInfo().IsDir() {
                continue
            }

            // Remove the top-level directory from the path
            // GitHub adds a directory like "owner-repo-commithash/"
            parts := strings.Split(file.Name, "/")
            if len(parts) > 1 {
                parts = parts[1:] // Skip the first directory
            }
            relativePath := path.Join(parts...)
            
            // Open the file from the zip
            rc, err := file.Open()
            if err != nil {
                continue
            }

            // Read the file contents
            content, err := io.ReadAll(rc)
            if err != nil {
                rc.Close()
                continue
            }
            rc.Close()

            // Store the file contents in memory
            fileMap[relativePath] = content
            fileList = append(fileList, relativePath)
        }

        // Create response with file information
        response := map[string]interface{}{
            "status": "success",
            "files": fileList,
            "fileCount": len(fileList),
            "repository": map[string]string{
                "name": project.Name,
                "url": project.Repository.CloneURL,
            },
        }

        // Send response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
      
    })
}

func main() {}