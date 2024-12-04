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
        fmt.Println("Archive URL: ", archiveURL)

        // Create custom transport that follows redirects
        transport := &http.Transport{
            Proxy: http.ProxyFromEnvironment,
            ForceAttemptHTTP2: true,
        }

        client := &http.Client{
            Transport: transport,
            CheckRedirect: func(req *http.Request, via []*http.Request) error {
                // Copy the authorization header to the redirected request
                req.Header.Set("Authorization", "Bearer "+authToken)
                return nil
            },
        }

        req, err := http.NewRequest("GET", archiveURL, nil)
        if err != nil {
            http.Error(w, "Failed to create request", http.StatusInternalServerError)
            return
        }

        req.Header.Set("Authorization", "Bearer "+authToken)
        req.Header.Set("Accept", "application/vnd.github+json")

        resp, err := client.Do(req)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to get archive: %v", err), http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        // Check response status
        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            http.Error(w, fmt.Sprintf("GitHub API error: %d - %s", resp.StatusCode, string(body)), http.StatusInternalServerError)
            return
        }

        // Rest of your code remains the same...
        zipBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            http.Error(w, "Failed to read zip file", http.StatusInternalServerError)
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
                continue
            }

            content, err := io.ReadAll(rc)
            if err != nil {
                rc.Close()
                continue
            }
            rc.Close()

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
        json.NewEncoder(w).Encode(response)
    })
}

func main() {}