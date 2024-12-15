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
	/* Deploying a Project is the Most Complex Operation in the System
	*  General High Level Process
	*  Receive a Project(user, repo, header) -> locate nvms/readme and codebase (Send to Provisioner Route)
	*  Unmarshal the NVMS(yaml) as an Object and Validate/Process it
	*  Begin Generating a Resource Plan -> Send to Builder Route 
	*  Build VPC/Network, Configure Security Groups, Setup Load Balancers
	*
	*
	*/

}