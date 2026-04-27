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

        zipResp, err := DownloadRepo(w, r, archiveURL, authToken)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        } 
        defer zipResp.Body.Close()
        fmt.Println("Processing zip file...")
        /
        zipReader, err := processZip(zipResp, w,r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Println("Reading zip file...")
        nvmsFile, readmeFile, err := readZip(zipReader)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Println("NVMS File: ", nvmsFile)
        fmt.Println("Readme File: ", readmeFile)
        response := map[string]interface{}{
            "status": "success",
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
func readFileFromZip(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
        fmt.Println("Error opening file: ", err)
		return "", err
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, rc); err != nil {
		return "", err
	}
	return buf.String(), nil
}
func readZip(r *zip.Reader) (string,string,error) {
    
        	var nvmsFile, readmeFile string
            var err error
	for _, f := range r.File {
		// Check for *.nvms file
		if strings.HasSuffix(f.Name, ".nvms") && nvmsFile == "" {
			nvmsFile, err = readFileFromZip(f)
			if err != nil {
				return "", "", err
			}
		}
		// Check for README.* file
		if strings.HasPrefix(strings.ToLower(f.Name), "readme.") && readmeFile == "" {
			readmeFile, err = readFileFromZip(f)
			if err != nil {
				return "", "", err
			}
		}
		// Break early if both files are found
		if nvmsFile != "" && readmeFile != "" {
			break
		}
	}
	return nvmsFile, readmeFile, nil

}
func processZip(zipResp *http.Response, w http.ResponseWriter, r *http.Request) (*zip.Reader, error) {
     zipBytes, err := io.ReadAll(zipResp.Body)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to read zip file: %v", err), http.StatusInternalServerError)
        
        }

        if len(zipBytes) == 0 {
            http.Error(w, "Received empty zip file", http.StatusInternalServerError)
           
        }

        zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to read zip archive: %v", err), http.StatusInternalServerError)
          
        }
        return zipReader, nil
}
func DownloadRepo(w http.ResponseWriter, r *http.Request, archiveURL string, authToken string) (*http.Response, error) {
    // Use Spin's HTTP client for the initial request
        req,err := http.NewRequest("GET", archiveURL, nil)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
         
        }
        req.Header.Add("User-Agent", "Byteport") // Set user agent  
        req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken)) // Add token to headers
        req.Header.Add("Accept", "application/vnd.github+json") // Set accept header to GitHub API
        //fmt.Println("HEAD: ", req.Header)
        // print out full request
        //fmt.Println("Request: ", req)
        resp, err := spinhttp.Send(req)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to get archive: %v", err), http.StatusInternalServerError)
        
        }
        defer resp.Body.Close()
        var zipResp *http.Response
        if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
            http.Error(w, fmt.Sprintf("Expected redirect or file, got: %s", resp.Status), resp.StatusCode)
           
        }
        
        zipResp = resp
        
        return zipResp, nil
}
func main() {}