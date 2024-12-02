package main

import (
	"backend/nvms/models"
	"fmt"
	"net/http"
)
func DeployProject(w http.ResponseWriter, r *http.Request) {
	var repository models.Repository;
	var project models.Project;
	r.Body.Read(repository);
	r.Body.Read(project);
	fmt.Println("Deploying project: ", project)

}