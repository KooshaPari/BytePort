package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"nvms/models"
)
func DeployProject(w http.ResponseWriter, r *http.Request) {
	var repository models.Repository;
	var project models.Project;
	body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body", http.StatusInternalServerError)
        return
    }
    defer r.Body.Close()

    err = json.Unmarshal(body, &data)
    if err != nil {
        http.Error(w, "Error parsing JSON", http.StatusBadRequest)
        return
    }
    project := data.Project

   
	fmt.Println("Deploying project: ", project)

}