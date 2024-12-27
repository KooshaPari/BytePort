package projectManager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nvms/lib"
	"nvms/models"
)




func TerminateProject(w http.ResponseWriter, r *http.Request) {
 /*Get Project, User from Req -> Deployments from DeploymentsJSON, loop thru call a terminate resource func(analyze service type choose appropriate termination function)*/
 var project models.Project; var user models.User;
 project, user, err := readBody(w, r)
 if err != nil {
	 return
 }
 var deployments map[string]models.Instance
 err = json.Unmarshal([]byte(project.DeploymentsJSON), &deployments)
 if err != nil {
	 http.Error(w, "Error parsing JSON", http.StatusBadRequest)
	 return
 }
 accessKey,secretKey, err := lib.GetAWSCredentials(user) 
 if err != nil {
	 http.Error(w, "Error getting AWS credentials", http.StatusInternalServerError)
	 return
 }
remainingDeployments := deployments
for len(remainingDeployments) > 0 {
	for deploymentName, deployment := range remainingDeployments {
		deploymentCompleted := true
		remainingResources := deployment.Resources
		var newResources []models.AWSResource
		for len(remainingDeployments) > 0 {
		for _, resource := range remainingResources {
			termCompleted := true
			// Try to terminate the resource
			err :=  terminateResource(w, r, resource, accessKey, secretKey)
			if err != nil {
				// If termination fails, mark deployment as incomplete
				fmt.Println("Error terminating resource: ", err)
				deploymentCompleted = false
				termCompleted  = false
				newResources = append(newResources, resource)
				continue
			}
			if termCompleted {
				fmt.Println("Termination Completed: ", resource.Name)
			}
			
		}
		// If all resources in deployment were terminated, remove it from remaining
		if deploymentCompleted {
			fmt.Println("Deployment Completed: ", deploymentName)
			delete(remainingDeployments, deploymentName)
		}
	}
}
w.WriteHeader(http.StatusOK)
 fmt.Println("Project Terminated")

}
func terminateResource(w http.ResponseWriter, r *http.Request, resource models.AWSResource, accessKey string, secretKey string) (error){
	 
	if resource.Type == "S3"{ 
		err := lib.TerminateS3(resource, accessKey, secretKey)
		if err != nil {
			http.Error(w, "Error terminating S3: "+err.Error(), http.StatusInternalServerError)
			return err
		}}
	if resource.Type == "EC2"{ 
		err := lib.TerminateEC2(resource, accessKey, secretKey)
		if err != nil {
			http.Error(w, "Error terminating EC2: "+err.Error(), http.StatusInternalServerError)
			return err
		}}
	if resource.Type == "ALB"{
		err := lib.TerminateALB(resource, accessKey, secretKey)
		if err != nil {
			http.Error(w, "Error terminating ALB: "+err.Error(), http.StatusInternalServerError)
			return err
		} }
	if resource.Type == "TargetGroup"{ 
		err := lib.TerminateTargetGroup(resource, accessKey, secretKey)
		if err != nil {
			http.Error(w, "Error terminating TargetGroup: "+err.Error(), http.StatusInternalServerError)
			return err
		}}
	 
		return nil
} 