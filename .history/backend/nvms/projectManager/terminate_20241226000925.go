package projectManager

import (
	"encoding/json"
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
 for _, deployment := range deployments {
	for _, resource := range deployment.Resources{
	 terminateResource(w,r,resource, accessKey, secretKey)}

 }

}
func terminateResource(w http.ResponseWriter, r *http.Request, resource models.AWSResource, accessKey string, secretKey string){
	switch resource.Type {
	case "S3":
		err := lib.TerminateS3(resource, accessKey, secretKey)
		if err != nil {
			http.Error(w, "Error terminating S3: "+err.Error(), http.StatusInternalServerError)
			return
		}
	case "EC2":
		err := lib.TerminateEC2(resource, accessKey, secretKey)
		if err != nil {
			http.Error(w, "Error terminating EC2: "+err.Error(), http.StatusInternalServerError)
			return
		}
	if resource.Type == "ALB":
		err := lib.TerminateALB(resource, accessKey, secretKey)
		if err != nil {
			http.Error(w, "Error terminating ALB: "+err.Error(), http.StatusInternalServerError)
			return
		}
	if resource.Type == "TargetGroup"{ 
		err := lib.TerminateTargetGroup(resource, accessKey, secretKey)
		if err != nil {
			http.Error(w, "Error terminating TargetGroup: "+err.Error(), http.StatusInternalServerError)
			return
		}}
	 

}}