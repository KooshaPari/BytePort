package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/models"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		var user models.User; var project models.Project;   
 
		project, err := getRequestDetails(w,r)
		if err != nil {
			fmt.Println("err getting dets: ", err)
			http.Error(w, "Error Reading Request", http.StatusInternalServerError)
			return
		}
		user = project.User
		// decrypt portDet
		 
		// pull Portfolio API Format
		// Generate Response (hand info above to ai langchain)
		// post to portfolio
		// return
		fmt.Fprintln(w, "Hello Fermyon!")
	})
}
func getRequestDetails(w http.ResponseWriter, r *http.Request)(models.Project,   error){
	fmt.Println("Getting Template Dets")
	var project models.Project
	body, err := io.ReadAll(r.Body)
        if err != nil {
            fmt.Println("Error reading request body: ", err)
            http.Error(w, "Error reading request body", http.StatusInternalServerError)
            return "",err
        }
        defer r.Body.Close()
	fmt.Println("Parsing JSON...")
        err = json.Unmarshal(body, &project)
        if err != nil {
            http.Error(w, "Error parsing JSON", http.StatusBadRequest)
            return "",err
        }
		return project, nil
}

func main() {}
