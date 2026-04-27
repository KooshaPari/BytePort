package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"nvms/models"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/google/go-github/v67/github"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		var project models.Project;
		body, err := ioutil.ReadAll(r.Body)
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
		repoURL := project.Repository.URL;
		userEncryptedToken := project.User.Git.Token;
		authToken, err := lib.DecryptSecret(userEncryptedToken)
		if err != nil {
			http.Error(w, "Failed to decrypt user access token", http.StatusInternalServerError)
			return
		}

		// decrypt token

		client := github.NewClient(nil).WithAuthToken(authToken)
		github.Downl
		
		// pull project from repo


	})
}

func main() {}
