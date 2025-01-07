package projectManager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/lib"
	ec2 "nvms/lib/awspin/ec2"
	"nvms/models"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
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

 func parseNVMS(yamlContent string) (*models.NVMS, error) {
	// send object to /parse
	var nvms *models.NVMS
 
	reqBody, err := json.Marshal(yamlContent)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", "/parse", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	resp, err := spinhttp.Send(req)
	if err != nil || http.StatusOK != resp.StatusCode {
		return nil, err
	}
	//unmarshal to nvms
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &nvms)
	if err != nil {
		return nil, err
	}
	return nvms, nil
}
 func ProvisionFiles(w http.ResponseWriter, r *http.Request,project models.Project)(string,string,[]byte,[]string,error){
	var nvmsString string
	var readMeString string
	var codebase []byte
	var files []string
	reqBody, err := json.Marshal(project)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return nvmsString,readMeString,codebase,files,err
	}
	req, err := http.NewRequest("GET", "/getter", bytes.NewReader(reqBody))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return nvmsString,readMeString,codebase,files,err
	}
	resp, err := spinhttp.Send(req)
	if err != nil || http.StatusOK != resp.StatusCode {
		http.Error(w, "Error sending request", http.StatusInternalServerError)
		return nvmsString,readMeString,codebase,files,err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response body", http.StatusInternalServerError)
		return nvmsString,readMeString,codebase,files,err
	}
	var response ProvisionerResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return nvmsString,readMeString,codebase,files,err
	}
	nvmsString = response.Nvms
	readMeString = response.Readme
	codebase = response.ZipBall
	files  = response.FileMap
	return nvmsString,readMeString,codebase,files,nil
 } 
 func addToDemo(project models.Project)(error){
	reqBody, err := json.Marshal(project)
	if err != nil {
		return fmt.Errorf("error marshaling project: %w", err)
	}
	req, err := http.NewRequest("GET", "/generate", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	_, err = spinhttp.Send(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	return nil
}
func DeployNVMSService(AccessKey string, SecretKey string, Bucket lib.S3DeploymentInfo, service models.Service, fileMap []string) ([]lib.EC2InstanceInfo, error) {
	instances, err := lib.DeployEC2(AccessKey, SecretKey, Bucket, service, fileMap )
	if err != nil {
		fmt.Println("Error deploying EC2: ", err)
		return nil, err
	}
	//fmt.Println("Deployed EC2 Instances: ", instances)
	fmt.Println("Building Services: ", service)

	return instances, nil
}
func DeployNVMSServiceMVM(AccessKey string, SecretKey string, Bucket lib.S3DeploymentInfo, service models.Service, fileMap []string, projName string, instProf ec2.CreateInstanceProfileResponse, targetRepo ec2.ContainerRepo, s3Info S3DeploymentInfo) ( error) {
	infraConfig, err := lib.CreateInfrastructureConfiguration(AccessKey, SecretKey, projName, instProf.InstanceProfile.InstanceProfileName)
	if err != nil {
		fmt.Println("Error creating infrastructure configuration: ", err)
		return err;
	}
	components, err := lib.CreateImageComponents(AccessKey, SecretKey, "/app"+service.Path, projName+"-"+service.Name,s3Info,service,fileMap)
	if err != nil {
		fmt.Println("Error creating image components: ", err)
		return err;
	}
	containerRecipe, err := lib.CreateContainerRecipe(AccessKey, SecretKey, "/app"+service.Path, projName+"-"+service.Name, components, targetRepo)
	if err != nil {
		fmt.Println("Error creating container recipe: ", err)
		return err;
	}
	pipeline, err := lib.CreateImgPipeline(AccessKey, SecretKey, projName, infraConfig.InfrastructureConfigurationArn)
	if err != nil {
		fmt.Println("Error creating image pipeline: ", err)
		return err;
	}
	buildVersion, err := lib.ExecuteImgPipeline(AccessKey, SecretKey, projName, pipeline.ImagePipelineArn)
	
	return nil
}