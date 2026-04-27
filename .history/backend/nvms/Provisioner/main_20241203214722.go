package main

import (
	"archive/zip"
	"bytes"
	"compress/flate"
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

func init() {
    // Register the deflate compression method
    zip.RegisterDecompressor(zip.Deflate, func(r io.Reader) io.ReadCloser {
        return flate.NewReader(r)
    })

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
            http.Error(w, "Failed to decrypt user access token: " + err.Error(), http.StatusInternalServerError)
            return
        }

        archiveURL := getArchiveURL(project.Repository.ArchiveURL)
        fmt.Println("Initial Archive URL: ", archiveURL)

        // Get the zip content
        zipBytes, err := fetchZipContent(archiveURL, authToken)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to fetch zip content: %v", err), http.StatusInternalServerError)
            return
        }

        if len(zipBytes) == 0 {
            http.Error(w, "Received empty zip file", http.StatusInternalServerError)
            return
        }

        // Create zip reader
        zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to read zip archive: %v", err), http.StatusInternalServerError)
            return
        }
        
        fileMap := make(map[string][]byte)
        var fileList []string
        
        for _, file := range zipReader.File {
            // Debug information
            fmt.Printf("File: %s, Method: %d\n", file.Name, file.Method)
            
            if file.FileInfo().IsDir() {
                continue
            }

            // Only process files with supported compression methods
            if file.Method != zip.Deflate && file.Method != zip.Store {
                fmt.Printf("Warning: Unsupported compression method %d for file %s\n", file.Method, file.Name)
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
                "url": project.Repository.ArchiveURL,
            },
        }

        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(response); err != nil {
            http.Error(w, "Failed to encode response", http.StatusInternalServerError)
            return
        }
    })
}

// fetchZipContent handles the HTTP requests and redirects
func fetchZipContent(archiveURL, authToken string) ([]byte, error) {
    req, err := http.NewRequest("GET", archiveURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %v", err)
    }
    
    req.Header.Add("User-Agent", "Byteport")
    req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
    req.Header.Add("Accept", "application/vnd.github+json")
    
    resp, err := spinhttp.Send(req)
    if err != nil {
        return nil, fmt.Errorf("failed to get archive: %v", err)
    }
    defer resp.Body.Close()

    // Handle redirect if necessary
    if resp.StatusCode == http.StatusFound {
        redirectURL := resp.Header.Get("Location")
        if redirectURL == "" {
            return nil, fmt.Errorf("no redirect URL provided")
        }

        redirReq, err := http.NewRequest("GET", redirectURL, nil)
        if err != nil {
            return nil, fmt.Errorf("failed to create redirect request: %v", err)
        }

        resp, err = spinhttp.Send(redirReq)
        if err != nil {
            return nil, fmt.Errorf("failed to download zip: %v", err)
        }
        defer resp.Body.Close()
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    return io.ReadAll(resp.Body)
}

func main() {}