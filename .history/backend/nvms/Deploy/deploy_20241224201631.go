package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/lib"
	"nvms/models"
	"strings"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
	"gopkg.in/yaml.v2"
)

type ProvisionerResponse struct {
	Nvms    string `json:"nvmsFile"`
	Readme  string `json:"Readme"`
	ZipBall []byte `json:"zipball"`
	FileMap []string `json:"fileMap"`
}

func parseNVMSConfig(yamlContent string) (*models.NVMS, error) {
    fmt.Printf("Parsing YAML content: %s\n", yamlContent) // Debug log
    
	config := &models.NVMS{
		Services: []models.Service{},
	}
    
    // Validate YAML content
    if strings.TrimSpace(yamlContent) == "" {
        return nil, fmt.Errorf("empty YAML content")
    }

    // Parse YAML with error handling
    err := yaml.Unmarshal([]byte(yamlContent), config)
    if err != nil {
        return nil, fmt.Errorf("YAML parsing error: %w", err)
    }

    // Validate required fields
    if config.Name == "" {
        return nil, fmt.Errorf("missing required field: NAME")
    }
    if len(config.Services) == 0 {
        return nil, fmt.Errorf("no services defined in YAML")
    }

    // Validate each service
	found := false
    for name, svc := range config.Services {
        if svc.Path == "" {
            return nil, fmt.Errorf("service %s missing PATH", name)
        }
        if svc.Port == 0 {
            return nil, fmt.Errorf("service %s missing PORT", name)
        }
		if svc.Name == "main"{
			if(found){
				return nil, fmt.Errorf("service main already defined", name)
			}else{
				found = true
			}
		}
    }
	if (!found){
		return nil, fmt.Errorf("service main not defined")
	}

    return config, nil
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

	var user models.User
	var project models.Project
	var nvmsString string
	var readMeString string
	var codebase []byte
    // sec 1
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
	 
	user = project.User
	// sec 2
	reqBody, err := json.Marshal(project)
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
	var response ProvisionerResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}
	fmt.Println("Got files");
	//ln(response)
	nvmsString = response.Nvms
	readMeString = response.Readme
	codebase = response.ZipBall
	files := response.FileMap
	 
	//TODO: Unmarshal the NVMS(yaml) as an Object and Validate/Process it
	 
	//fmt.Println("Project: ", project)
	 
	fmt.Println("ReadMe: ", readMeString)
	//fmt.Println("Files: ", files) 
	//fmt.Println("Codebase: ", codebase)
	
	nvmsConfig, err := parseNVMSConfig(nvmsString)
	if err!= nil{
		fmt.Println("Error parsing NVMS: ", err)
		http.Error(w, "Error parsing NVMS: "+err.Error(), http.StatusBadRequest)
	}
	project.NvmsConfig = *nvmsConfig
	project.Readme = readMeString
	// sec 3
	eAccKey := user.AwsCreds.AccessKeyID
	eSecKey := user.AwsCreds.SecretAccessKey

	accesskey, err := lib.DecryptSecret(eAccKey)
	if err != nil {
		http.Error(w, "Error decrypting secret: "+err.Error(), http.StatusInternalServerError)
		return
	}
	secretkey, err := lib.DecryptSecret(eSecKey)
	if err != nil {
		http.Error(w, "Error decrypting secret: "+err.Error(), http.StatusInternalServerError)
		return
	}

	bucket, err := lib.PushToS3(codebase, accesskey, secretkey, project.Name)
	 
	if err != nil {
		fmt.Println("Error pushing to S3: ", err)
		http.Error(w, "Error pushing to S3: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	 
	ServiceInstances := make(map[string][]lib.EC2InstanceInfo)
	 serviceMap  := make(map[string]models.Service )
	for _, service := range nvmsConfig.Services {
		/* So here we need to deploy 2x services into a vm, for now just a basic ec2, that will then be run with X command and have the following ports opened publicly for it. */
		fmt.Println("Serve")
		instances, err := DeployNVMSService(accesskey, secretkey, bucket, service,files)
		if err != nil {
			fmt.Println("Error deploying service: ", err)
			http.Error(w, "Error deploying service: "+err.Error(), http.StatusInternalServerError)
			return
		}
		

		serviceMap[service.Name] = service
		ServiceInstances[service.Name] = instances
	}
	fmt.Println("Handling Net...")
	
	
	fmt.Println("Service Instances: ", ServiceInstances)
	// wind down services since this is for debug.
	// await deployment then setup network
	// write service Instances to bodyservBody, err = json.Marshal(ServiceInstances)
	// run a for loop on service instances, for each await deploy then register services, prov network and finally listener rules.
	var instIDs []string
	fmt.Println("Waiting for Initialization")
	// add short wait
	fmt.Println("Initializing")
	for name, instances := range ServiceInstances {
		fmt.Println("Initializing ids: ", name)
		instIDs = []string{}
		for _, instance := range instances {
			instIDs = append(instIDs, instance.InstanceID)
		}
		fmt.Println("Initializing: ", name)
	err := lib.AwaitInitialization(accesskey, secretkey, instIDs)

	if err != nil {
		http.Error(w, "Error Checking init", http.StatusBadRequest)
		return
	}

	fmt.Println("Intialized: ", name)
	}
	
	lbArn, vpcId,accessURL, err := lib.ProvisionNetwork(accesskey, secretkey, project.Name )
	project.AccessURL = accessURL
	if err != nil {
		fmt.Println("Error provisioning network", err)
		http.Error(w, "Error provisioning network: "+err.Error(), http.StatusInternalServerError)
	}
	var listenArn string
	for _, instance := range ServiceInstances["main"]{
		fmt.Println("building main listener(s)")
		listenArn, err = lib.CreateALBListener(accesskey, secretkey, project.Name, lbArn, vpcId, instance.InstanceID,  serviceMap["main"].Port)
		fmt.Println("Built main listener(s) for instance: ", instance.InstanceID)
	}

	priority := 1
	for name, instances := range ServiceInstances {
		service := serviceMap[name]
		if(name != "main"){
		for _, instance := range instances {
			
		tgArn, err := lib.RegisterService(accesskey, secretkey, lbArn, project.Name, name, vpcId, instance.InstanceID,   service.Port )
		if err != nil {
			fmt.Println("Error registering service: ", err)
			http.Error(w, "Error registering service: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("registered service")
		err = lib.SetListenerRules(accesskey, secretkey, listenArn, tgArn, name, priority)
		if err != nil {
			fmt.Println("Error creating listener rule: ", err)
			http.Error(w, "Error creating listener rule: "+err.Error(), http.StatusInternalServerError)
			return
		}
		priority++
		fmt.Println("Created Listener Rule")
	}}

	}
    fmt.Println("Completed EC2-Deploy.")
	err = addToDemo(project)
	if err != nil {
		fmt.Println("error generating demo: ", err)
		http.Error(w,"error generating demo"+err.Error(), http.StatusInternalServerError)
	}
	projectJSON, err := json.Marshal(project)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(projectJSON) 
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
 

	 