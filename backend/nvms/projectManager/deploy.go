package projectManager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"nvms/lib"
	"nvms/models"
	"strings"
	"time"

github.com/google/uuid
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

	
	
    // sec 1
	project,user,err := readBody(w,r)
	if err != nil {
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}
	nvmsString, readMeString, codebase, _, err := ProvisionFiles(w, r, project)
	if err != nil {
		http.Error(w, "Error provisioning files", http.StatusInternalServerError)	
	}
	fmt.Println("Got files");
	//ln(response)
	
	// add new deployment to project
	 
	//TODO: Unmarshal the NVMS(yaml) as an Object and Validate/Process it
	 
	//fmt.Println("Project: ", project)
	 
	fmt.Println("ReadMe: ", readMeString)
	if project.GetDeploys()== nil {
        project.CreateDeploys()
    }

    deployID := uuid.New().String()
    project.AppendDeploy(deployID,models.Instance{
        UUID:            deployID,
        Name:            "main",
        Status:          "initializing",
        Owner:           user.UUID,
         
        Resources:       make([]models.AWSResource, 0), // Initialize slice
    })   
	//fmt.Println("Files: ", files) 
	//fmt.Println("Codebase: ", codebase)
	
	nvmsConfig, err := parseNVMSConfig(nvmsString)
	if err!= nil{
		fmt.Println("Error parsing NVMS: ", err)
		http.Error(w, "Error parsing NVMS: "+err.Error(), http.StatusBadRequest)
	}
	project.NvmsConfig = *nvmsConfig
	project.Readme = readMeString

	// Store project files locally (replaces S3)
	storageManager, err := lib.GetStorageManager()
	if err != nil {
		http.Error(w, "Error initializing storage", http.StatusInternalServerError)
		return
	}

	localStorage, err := storageManager.PushToLocalStorage(codebase, project.Name)
	if err != nil {
		fmt.Println("Error storing project locally: ", err)
		http.Error(w, "Error storing project locally: "+err.Error(), http.StatusInternalServerError)
		return
	}

	instance := project.GetDeploy(deployID)
	instance.Resources = append(instance.Resources, models.AWSResource{
		Name:    "Local-Storage",
		ARN:     localStorage.BucketARN,
		Status:  "deployed",
		Region:  localStorage.Region,
		ID:      localStorage.BucketName,
		Type:    "LocalStorage",
		Service: "general",
	})
	project.AppendDeploy(deployID, instance)
     

	// Deploy services using Docker (replaces EC2)
	dockerManager, err := lib.GetDockerManager()
	if err != nil {
		http.Error(w, "Error initializing Docker manager", http.StatusInternalServerError)
		return
	}

	serviceInstances := make(map[string][]lib.DockerInstanceInfo)
	serviceMap := make(map[string]models.Service)

	for _, service := range nvmsConfig.Services {
		fmt.Println("Deploying service:", service.Name)

		// Set project name for service
		service.ProjectName = project.Name

		// Deploy service in Docker container
		containerInfo, err := dockerManager.CreateAndStartContainer(service, localStorage.Path)
		if err != nil {
			fmt.Println("Error deploying service:", err)
			http.Error(w, "Error deploying service: "+err.Error(), http.StatusInternalServerError)
			return
		}

		serviceInstances[service.Name] = []lib.DockerInstanceInfo{*containerInfo}
		serviceMap[service.Name] = service

		// Update deployment resources
		res := project.GetDeploy(deployID)
		res.Resources = append(res.Resources, models.AWSResource{
			Name:    service.Name + "-Container",
			ARN:     containerInfo.ContainerID,
			Status:  "deployed",
			Region:  "local",
			ID:      containerInfo.ContainerID,
			Type:    "Docker",
			Service: service.Name,
		})
		project.AppendDeploy(deployID, res)
	}
	fmt.Println("Handling Network...")
	fmt.Println("Service Instances: ", serviceInstances)

	// Wait for containers to be ready
	fmt.Println("Waiting for containers to initialize...")
	err = waitForContainersReady(serviceInstances)
	if err != nil {
		http.Error(w, "Error waiting for containers: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("All containers initialized")
	
	// Create Cloudflare tunnel (replaces ALB/Load Balancer)
	fmt.Println("Setting up tunnel...")
	tunnelManager, err := lib.GetTunnelManager()
	if err != nil {
		http.Error(w, "Error initializing tunnel manager", http.StatusInternalServerError)
		return
	}

	// Convert services for tunnel configuration
	var tunnelServices []models.Service
	for _, service := range nvmsConfig.Services {
		tunnelServices = append(tunnelServices, service)
	}

	_, err = tunnelManager.CreateProjectTunnel(project.Name, tunnelServices)
	if err != nil {
		fmt.Println("Error creating tunnel:", err)
		http.Error(w, "Error creating tunnel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Start the tunnel
	actualURL, err := tunnelManager.StartProjectTunnel(project.Name)
	if err != nil {
		fmt.Println("Error starting tunnel:", err)
		http.Error(w, "Error starting tunnel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	project.AccessURL = actualURL

	// Update deployment with tunnel info
	res := project.GetDeploy(deployID)
	res.Resources = append(res.Resources, models.AWSResource{
		Name:    "Cloudflare-Tunnel",
		ARN:     actualURL,
		Status:  "deployed",
		Region:  "global",
		ID:      project.Name,
		Type:    "Tunnel",
		Service: "general",
	})
	project.AppendDeploy(deployID, res)
	 
		if !strings.HasPrefix(project.AccessURL, "http") {
			project.AccessURL = "http://" + project.AccessURL	
		}
    fmt.Println("Completed EC2-Deploy.")
	fmt.Println("Project: ", project) 
	if err := project.BeforeSave(); err != nil {
		http.Error(w, "Error saving project", http.StatusInternalServerError)}
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
// waitForContainersReady waits for all containers to be in running state
func waitForContainersReady(serviceInstances map[string][]lib.DockerInstanceInfo) error {
	dockerManager, err := lib.GetDockerManager()
	if err != nil {
		return fmt.Errorf("failed to get Docker manager: %w", err)
	}

	for serviceName, instances := range serviceInstances {
		for _, instance := range instances {
			for i := 0; i < 30; i++ { // Wait up to 30 seconds
				status, err := dockerManager.GetContainerStatus(instance.ContainerID)
				if err != nil {
					return fmt.Errorf("failed to get container status for %s: %w", serviceName, err)
				}

				if status == "running" {
					break
				}

				if i == 29 {
					return fmt.Errorf("container %s failed to start within timeout", serviceName)
				}

				// Wait 1 second before checking again
				time.Sleep(time.Second)
			}
		}
	}

	return nil
}
 

	 