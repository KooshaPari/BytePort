package main

import (
	"fmt"
	"net/http"

	"C:/Users/koosh/Dev/Bytep/nvms/models"
)
func DeployProject(w http.ResponseWriter, r *http.Request) {
	var repository models.Repository;
	var project models.Project;
	r.Body.Read(repository);
	r.Body.Read(project);
	fmt.Println("Deploying project: ", project)

}