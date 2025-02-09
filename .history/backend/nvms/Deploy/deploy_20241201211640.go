package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"nvms/models"
)
func DeployProject(w http.ResponseWriter, r *http.Request) {
	//var repository models.Repository;
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


   
	fmt.Println("Deploying project: ", project)
	// forward request to /deploy 
	http.NewRequest("POST", "http://localhost:3000/deploy", bytes.NewBuffer(body))

}