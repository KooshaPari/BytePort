package main

import (

	"net/http"
	"backend/nvms/models"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)
func DeployProject(w http.ResponseWriter, r *http.Request) {
	var repository models.Repository;
	var project models.Project;
	r.Body.Read(repository);
	r.Body.Read(project);
	fmt.Println("Deploying project: ", project)

}