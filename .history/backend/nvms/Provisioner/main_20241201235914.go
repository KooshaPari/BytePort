package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/lib"
	"nvms/models"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
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

		// decrypt token
		
		if err != nil {
			http.Error(w, "Failed to clone repository", http.StatusInternalServerError)
			return
		}
		fmt.Println("Cloned repository: ", repo)
		

		// pull project from repo


	})
}

func main() {}
