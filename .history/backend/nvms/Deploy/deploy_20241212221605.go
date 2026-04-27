package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/models"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
	"gopkg.in/yaml.v3"
)
type ProvisionerResponse struct {
	Nvms string `json:"nvmsFile"`
	Readme string `json:"readmeFile"`
	ZipBall []byte `json:"zipball"`
}
func DeployProject(w http.ResponseWriter, r *http.Request) {
	/*  Deploying a Project is the Most Complex Operation in the System
	*   General High Level Process
	*   Receive a Project(user, repo, header) -> 
		locate nvms/readme and codebase (Send to Provisioner Route)
	*   Unmarshal the NVMS(yaml) as an Object and Validate/Process it
	*   Begin Generating a Resource Plan -> Send to Builder Route 
	*   Build VPC/Network, Configure Security Groups, Setup Load  *  *   Balancers, Go down the line of the Resource Plan/NVMS Object
	*   Validate Resources and Send Status -> Deployment Module
	*   Config/Deploy MicroVM(FireCracker), Config Services, Setup 	  
	*   Monitoring call portfolio route (Repository, Readme, NVMS)
	*   Analyze Project (Get Details for Prompting, Read Playground Type from NVMS), Pull Templates from Portfolio, Pick appropriate template given args and build and send back.
	*   Open appropriate connections for playground and rpovide to route, deployed.
	*/
	var user models.User; var project models.Project; var nvmsString string; var readMeString string; var codebase []byte;

	body, err := io.ReadAll(r.Body)
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
	//fmt.Println(project)
	user = project.User
	reqBody,err := json.Marshal(project)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	req, err := http.NewRequest("GET", "/provision", bytes.NewReader(reqBody))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}
	resp, err := spinhttp.Send(req)
	if err != nil || http.StatusOK != resp.StatusCode {
		http.Error(w, "Error sending request", http.StatusInternalServerError)
		return
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response body", http.StatusInternalServerError)
		return
	}
	var response ProvisionerResponse;
	err = json.Unmarshal(body, &response)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	//ln(response)
	nvmsString = response.Nvms
	readMeString = response.Readme
	codebase = response.ZipBall
	//TODO: Unmarshal the NVMS(yaml) as an Object and Validate/Process it
	fmt.Println("User: ", user)
	fmt.Println("Project: ", project)
	fmt.Println("NVMS: ", nvmsString)
	fmt.Println("ReadMe: ", readMeString)
	//fmt.Println("Codebase: ", codebase)
	nvmsConfig := models.NVMSConfig{
		
	}
	yaml.Unmarshal()
	//TODO: Begin Generating a Resource Plan -> Send to Builder Route
	//TODO: Build VPC/Network, Configure Security Groups, Setup Load Balancers, Go down the line of the Resource Plan/NVMS Object
	//TODO: Validate Resources and Send Status -> Deployment Module
	//TODO: Config/Deploy MicroVM(FireCracker), Config Services, Setup Monitoring call portfolio route (Repository, ReadMe, NVMS)
	//TODO: Analyze Project (Get Details for Prompting, Read Playground Type from NVMS), Pull Templates from Portfolio, Pick appropriate template given args and build and send back.
	//TODO: Open appropriate connections for playground and rpovide to route, deployed.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Deploying Project"))
}


