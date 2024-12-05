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

func getArchiveURL(archiveURL string) string {
    fmt.Println("Archive URL: ", archiveURL)
    splitURL := strings.Split(archiveURL, "/{")
    if len(splitURL) > 1 {
        return splitURL[0] + "/zipball"
    }
    return archiveURL
}

func init() {
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

        
        fmt.Println("Reading zip file...")
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

func DownloadRepo() (*http.Response, error) {
    
}
func main() {}