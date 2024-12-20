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
	"strings"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

 
func init() {
    spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Received request")
        InitializeZipManually()
        var project models.Project
        fmt.Println("Reading request body...")
        body, err := io.ReadAll(r.Body)
        if err != nil {
            fmt.Println("Error reading request body: ", err)
            http.Error(w, "Error reading request body", http.StatusInternalServerError)
            return
        }
        defer r.Body.Close()
        fmt.Println("Parsing JSON...")
        err = json.Unmarshal(body, &project)
        if err != nil {
            http.Error(w, "Error parsing JSON", http.StatusBadRequest)
            return
        }
        fmt.Println("Decoding user access token...")
        authToken, err := lib.DecryptSecret(project.User.Git.Token)
        if err != nil {
            fmt.Println("Failed to decrypt user access token: ", err)
            http.Error(w, "Failed to decrypt user access token: " + err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Println("Getting archive URL...") 
        archiveURL := getArchiveURL(project.Repository.ArchiveURL)
        fmt.Println("Initial Archive URL: ", archiveURL)

        zipResp, err := DownloadRepo(w, r, archiveURL, authToken)
        if err != nil {
            fmt.Println("Error downloading zip file: ", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        } 
        fmt.Println("Processing zip file...")



        nvmsFileContent, readMeContent,projectZip, err := processZip(zipResp, w,r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Println("Extracted files")
     
        if(nvmsFileContent == "" || readMeContent == "" || projectZip == nil){
            http.Error(w, "Failed to extract files", http.StatusInternalServerError)
            return
        }
        // Parse Validate NVMS


        // push to s3


        response := map[string]interface{}{
            "NVMSFile": nvmsFileContent,
            "README": readMeContent,
            "ZipBall": projectZip,
        }

        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(response); err != nil {
            http.Error(w, "Failed to encode response", http.StatusInternalServerError)
            return
        }
    })
}
 
func main() {}
