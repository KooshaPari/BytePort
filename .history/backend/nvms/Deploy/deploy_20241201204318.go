package main

import (
	"fmt"
	"net/http"

	"nvms/models"
)
func DeployProject(w http.ResponseWriter, r *http.Request) {
	var repository models.Repository;
	var project models.Project;
	r.Body.Read(project);
	fmt.Println("Deploying project: ", project)

}