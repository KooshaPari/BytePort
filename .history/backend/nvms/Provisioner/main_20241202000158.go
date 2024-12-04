package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/lib"
	"nvms/models"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/hashicorp/go-getter"

	
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		var project models.Project;
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
		repoURL := project.Repository.CloneURL;
		userEncryptedToken := project.User.Git.Token;
		authToken, err := lib.DecryptSecret(userEncryptedToken)
		if err != nil {
			http.Error(w, "Failed to decrypt user access token", http.StatusInternalServerError)
			return
		}
		 sourceURL := fmt.Sprintf("git::https://oauth2:%s@%s",
        authToken,
        repo.GetCloneURL()[8:]) 
		 client := &getter.Client{
        Src:  sourceURL,
        Dst:  destPath,
        Mode: getter.ClientModeDir,
    }client.Get()
		if err != nil {
			http.Error(w, "Failed to clone repository", http.StatusInternalServerError)
			return
		}
		fmt.Println("Cloned repository: ", repo)
		

		// pull project from repo


	})
}

func main() {}
