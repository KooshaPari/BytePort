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
		
		var user models.User; var project models.Project;   
 
		project, err := getRequestDetails(w,r)
		if err != nil {
			fmt.Println("err getting dets: ", err)
			http.Error(w, "Error Reading Request", http.StatusInternalServerError)
			return
		}
		user = project.User
		// decrypt portDets
		decryptedPortEndpoint, err := lib.DecryptSecret(user.Portfolio.RootEndpoint)
		if err != nil {
			fmt.Println("Error decrypting endpoint:", err)
			http.Error(w, "Error decrypting endpoint", http.StatusInternalServerError)
			return
		}
		decryptedPortKey, err := lib.DecryptSecret(user.Portfolio.APIKey)
		if err != nil {
			fmt.Println("Error decrypting key:", err)
			http.Error(w, "Error decrypting key", http.StatusInternalServerError)
			return
		}
		 templateStruct, err := getTemplate(decryptedPortEndpoint, decryptedPortKey)
		// pull Portfolio API Format
		genRequest, err := generatePrompt(templateStruct, project)
		if err != nil{
			fmt.Println("error generating prompt: ", err)
			http.Error(w,"error generating prompt: ",http.StatusInternalServerError)
			return
		}
		// Generate Response (hand info above to ai langchain)
		// post to portfolio
		// return\
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello Fermyon!")
	})
}
func generatePrompt(template string, project models.Project) (string,error){
	base := `Given The Base User Inf`
	prompt := fmt.Sprintf(base,etcetc)
}
func getTemplate(endpoint string, key string)(string,error){
	uri := endpoint+"byteport"
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		fmt.Println("Error building request : ", err)
            
            return "",err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := spinhttp.Send(req)
	if err != nil {
		fmt.Println("Error sending request: ", err)
		   
			return "",err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body: ", err)
		return "", err
	}
	templateStruct, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Error marshaling JSON: ", err)
		return "", err
	}
	return string(templateStruct), nil

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
