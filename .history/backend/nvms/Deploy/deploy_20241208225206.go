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
	var user models.User; var project models.Project; var nvmsString string; var readMeString string;

	body, err := ioutil.ReadAll(r.Body)

}