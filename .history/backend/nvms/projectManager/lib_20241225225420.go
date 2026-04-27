package projectManager

import (
	"encoding/json"
	"io"
	"net/http"
	"nvms/models"
)
type ProvisionerResponse struct {
	Nvms    string `json:"nvmsFile"`
	Readme  string `json:"Readme"`
	ZipBall []byte `json:"zipball"`
	FileMap []string `json:"fileMap"`
}
func readBody(w http.ResponseWriter, r *http.Request)(models.Project,models.User,error){
	var user models.User
	var project models.Project
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return project, user, err
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &project)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return project, user, err
	}
	 
	user = project.User
	// sec 2
	if err := project.BeforeSave(); err != nil {
		http.Error(w, "Error saving project", http.StatusInternalServerError)
		return project, user, err
	}
	return project, user, nil
 } 
 func ProvisionFiles(w http.ResponseWriter, r *http.Request)(models.Project,models.User,error){
	var user models.User
	var project models.Project
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return project, user, err
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &project)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return project, user, err
	}
	 
	user = project.User
	// sec 2
	if err := project.BeforeSave(); err != nil {
		http.Error(w, "Error saving project", http.StatusInternalServerError)
		return project, user, err
	}
	return project, user, nil
 } 