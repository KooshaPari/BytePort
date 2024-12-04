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

        authToken, err := lib.DecryptSecret(project.User.Git.Token)
        if err != nil {
            http.Error(w, "Failed to decrypt user access token", http.StatusInternalServerError)
            return
        }

        archiveURL := getArchiveURL(project.Repository.ArchiveURL)
        fmt.Println("Initial Archive URL: ", archiveURL)

        // Use Spin's HTTP client for the initial request
        headers := map[string][]string{
            "Authorization": {fmt.Sprintf("Bearer %s", authToken)},
            "Accept":       {"application/vnd.github+json"},
            "User-Agent":   {"Go-Client"},
        }

        resp, err := spinhttp.Send(spinhttp.NewRequest("GET", archiveURL, headers))
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to get archive: %v", err), http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        // Check for redirect
        if resp.StatusCode != http.StatusFound {
            http.Error(w, fmt.Sprintf("Expected redirect, got: %s", resp.Status), resp.StatusCode)
            return
        }

        // Get redirect URL from Location header
        redirectURL := resp.Header.Get("Location")
        if redirectURL == "" {
            http.Error(w, "No redirect URL provided", http.StatusInternalServerError)
            return
        }

        fmt.Println("Redirect URL: ", redirectURL)

        // Make second request to the redirect URL using Spin's HTTP client
        zipResp, err := spinhttp.Get(redirectURL, nil))
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to download zip: %v", err), http.StatusInternalServerError)
            return
        }
        defer zipResp.Body.Close()

        if zipResp.StatusCode != http.StatusOK {
            http.Error(w, fmt.Sprintf("Failed to download zip, status: %s", zipResp.Status), zipResp.StatusCode)
            return
        }

        zipBytes, err := io.ReadAll(zipResp.Body)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to read zip file: %v", err), http.StatusInternalServerError)
            return
        }

        if len(zipBytes) == 0 {
            http.Error(w, "Received empty zip file", http.StatusInternalServerError)
            return
        }

        zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to read zip archive: %v", err), http.StatusInternalServerError)
            return
        }

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

        response := map[string]interface{}{
            "status": "success",
            "files": fileList,
            "fileCount": len(fileList),
            "repository": map[string]string{
                "name": project.Name,
                "url": project.Repository.CloneURL,
            },
        }

        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(response); err != nil {
            http.Error(w, "Failed to encode response", http.StatusInternalServerError)
            return
        }
    })
}

func main() {}