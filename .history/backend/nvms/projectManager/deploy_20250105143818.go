package projectManager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nvms/lib"
	ec2 "nvms/lib/awspin/ec2"
	"nvms/models"

	"github.com/google/uuid"
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
	nvmsString, readMeString, codebase, files, err := ProvisionFiles(w, r, project)
	if err != nil {
		http.Error(w, "Error provisioning files", http.StatusInternalServerError)	
	}
	fmt.Println("Got files");
	if project.GetDeploys()== nil {
        project.CreateDeploys()
    }
	fmt.Println("Creating deploy...")
    deployID := uuid.New().String()
    project.AppendDeploy(deployID,models.Instance{
        UUID:            deployID,
        Name:            "main",
        Status:          "initializing",
        Owner:           user.UUID,
         
        Resources:       make([]models.AWSResource, 0), // Initialize slice
    })   
	fmt.Println("parsing config...")
	nvmsConfig, err := parseNVMS(nvmsString)
	if err != nil{
		fmt.Println("Error parsing NVMS: ", err)
		http.Error(w, "Error parsing NVMS: "+err.Error(), http.StatusBadRequest)
	}
	project.NvmsConfig = *nvmsConfig
	project.Readme = readMeString
	// sec 3
	fmt.Println("getting creds...")
	 accesskey,secretkey, err := lib.GetAWSCredentials(user) 
	if err != nil {
		http.Error(w, "Error getting AWS credentials", http.StatusInternalServerError)
		return
	}
	bucket, instance,  ECRinfo, err := lib.PrepareProject(accesskey, secretkey, project.Name, codebase, files, nvmsConfig)
  	instance := project.GetDeploy(deployID)
    instance.Resources = append(instance.Resources, models.AWSResource{
        Name:    "S3-CodeBase Store",
        ARN:     bucket.BucketARN,
        Status:  "deployed",
        Region:  bucket.Region,
        ID:      bucket.BucketName,
		Type:   "S3",
            Service: "general",
        })
        project.AppendDeploy(deployID, instance)
		ecrRepo,err := lib.CreateECRRepo(accesskey, secretkey, project.Name)
		if err != nil {
			fmt.Println("Error creating ECR Repo: ", err)
			http.Error(w, "Error creating ECR Repo: "+err.Error(), http.StatusInternalServerError)
	 
			return
		}
		instance.Resources = append(instance.Resources, models.AWSResource{
			Name:    ecrRepo.Repository.RepositoryName,
			ARN:     ecrRepo.Repository.RepositoryArn,
			Status:  "deployed",
			Region:  ecrRepo.Repository.RepositoryUri,
			ID:      ecrRepo.Repository.RepositoryArn,
			Type:    "ECR",
			Service: "general",
		})
		project.AppendDeploy(deployID, instance)
		instanceProfile, err := lib.CreateInstanceProfile(accesskey, secretkey, project.Name)
		if err != nil {
			fmt.Println("Error creating instance profile: ", err)
			http.Error(w, "Error creating instance profile: "+err.Error(), http.StatusInternalServerError)
			return}
		instance.Resources = append(instance.Resources, models.AWSResource{
			Name:    "InstanceProfile",
			ARN:     instanceProfile.InstanceProfile.Arn,
			Status:  "deployed",
			Region:  instanceProfile.InstanceProfile.InstanceProfileName,
			ID:      instanceProfile.InstanceProfile.Arn,
			Type:    "InstanceProfile",
			Service: "general",
		})
		project.AppendDeploy(deployID, instance)
		//ServiceInstances := make(map[string][]lib.EC2InstanceInfo)
    	//serviceMap := make(map[string]models.Service)
		ECRinfo := ec2.ContainerRepo{
			RepositoryName: ecrRepo.Repository.RepositoryName,
			Service: 	  "general",
		}
		for _, service := range nvmsConfig.Services {
		fmt.Println("Serve")
		  err := DeployNVMSServiceMVM(accesskey, secretkey, bucket, service,files, project.Name, *instanceProfile,  ECRinfo)
		if err != nil {
			fmt.Println("Error deploying service: ", err)
			http.Error(w, "Error deploying service: "+err.Error(), http.StatusInternalServerError)
			return
		}

	}


	/*
	for _, service := range nvmsConfig.Services {
		fmt.Println("Serve")
		instances, err := DeployNVMSService(accesskey, secretkey, bucket, service,files)
		if err != nil {
			fmt.Println("Error deploying service: ", err)
			http.Error(w, "Error deploying service: "+err.Error(), http.StatusInternalServerError)
			return
		}
		res  := project.GetDeploy(deployID)
		var instanceIDs []string
		for _, inst := range instances {
			instanceIDs = append(instanceIDs, inst.InstanceID)
		}
		res.Resources = append(project.GetDeploy(deployID).Resources, models.AWSResource{
			Name: service.Name+"-Deployment",
			ARN: instances[0].InstanceID,
			Status: "deployed", 
			Region: instances[0].Region,
			ID: strings.Join(instanceIDs, ","),
			Type: "EC2",
			Service: service.Name,
		})
	project.AppendDeploy(deployID, res)
		serviceMap[service.Name] = service
		ServiceInstances[service.Name] = instances
	}
	fmt.Println("Handling Net...")
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
	
	alb, vpcId,accessURL, err := lib.ProvisionNetwork(accesskey, secretkey, project.Name )
	lbArn := alb.CreateLoadBalancerResult.LoadBalancers.Member.LoadBalancerArn
	res  := project.GetDeploy(deployID)
		res.Resources = append(project.GetDeploy(deployID).Resources, models.AWSResource{
		Name: "ALB",
		ARN: lbArn,
		Status: "deployed",
		Region: "us-east-1", 
		ID: lbArn, 
		Type: "ALB",
		Service: "general",
	 
	} )
	project.AppendDeploy(deployID, res) 

	project.AccessURL = accessURL
	if err != nil {
		fmt.Println("Error provisioning network", err)
		http.Error(w, "Error provisioning network: "+err.Error(), http.StatusInternalServerError)
	}
	var listenArn, targetGArn string
	// Create listener only for the first main instance
	if len(ServiceInstances["main"]) > 0 {
		instance := ServiceInstances["main"][0]
		fmt.Println("building main listener(s)")
		listenArn,targetGArn, err  = lib.CreateALBListener(accesskey, secretkey, project.Name, lbArn, vpcId, instance.InstanceID,  serviceMap["main"].Port)
 
		if err != nil {
		fmt.Println("Error Creating Listener  ", err)
		http.Error(w, "Error Creating Listener  : "+err.Error(), http.StatusInternalServerError)
		}

		res  := project.GetDeploy(deployID)
		res.Resources = append(project.GetDeploy(deployID).Resources, models.AWSResource{
		Name: "ALBListener",
		ARN: listenArn,
		Status: "deployed",
		Region: "us-east-1", 
		ID: listenArn, 
		Type:"Listener", 
		Service: "general",
		
	})
	res.Resources = append(project.GetDeploy(deployID).Resources, models.AWSResource{
		Name: "TargetGroup",
		ARN: targetGArn,
		Status: "deployed",
		Region: "us-east-1", 
		ID: targetGArn, 
		Type:"TargetGroup",
		Service: "main",
		
	})
	project.AppendDeploy(deployID, res) 
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
		res  := project.GetDeploy(deployID)
		res.Resources = append(project.GetDeploy(deployID).Resources, models.AWSResource{
		Name: service.Name+"-TargetGroup",
		ARN: tgArn,
		Status: "deployed",
		Region: "us-east-1", 
		Type: "TargetGroup",
		ID: tgArn, 
		Service: name,
 
	 
	} )
	project.AppendDeploy(deployID, res) 
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
	if !strings.HasPrefix(project.AccessURL, "http") {
		project.AccessURL = "http://" + project.AccessURL	
	}
    fmt.Println("Completed EC2-Deploy.")
	if err := project.BeforeSave(); err != nil {
		http.Error(w, "Error saving project", http.StatusInternalServerError)}
	err = addToDemo(project)
	if err != nil {
		fmt.Println("error generating demo: ", err)
		http.Error(w,"error generating demo"+err.Error(), http.StatusInternalServerError)
	}*/
	projectJSON, err := json.Marshal(project)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(projectJSON) 
}

 

	 