package main

import (
	"fmt"
	"net/http"
	"nvms/models"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		var user models.User; var project models.Project; var nvms models.NVMS; var readMe string
 
		project, readMe, err := getRequestDetails()
		if err != nil {

		}
		// receive Project Obj, Readme, NVMS 
		// pull Portfolio API Format
		// Generate Response (hand info above to ai langchain)
		// post to portfolio
		// return
		fmt.Fprintln(w, "Hello Fermyon!")
	})
}
func getRequestDetails(w http.ResponseWriter, r *http.Request)(models.Project,  string,error){
	fmt.Println("Getting Template Dets")
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
}

func main() {}
