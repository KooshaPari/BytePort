package main

import (
	"net/http"
	"nvms/models"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	google go github
	""
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
		// pull project from repo


	})
}

func main() {}
