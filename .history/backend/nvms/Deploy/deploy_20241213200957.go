package deploy

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nvms/lib"
	"nvms/models"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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
	nvmsConfig := lib.NVMS{}
	err = yaml.Unmarshal([]byte(nvmsString), &nvmsConfig)
	if err != nil {
		http.Error(w, "Error parsing YAML", http.StatusBadRequest)
		return
	}
	eAccKey := user.AwsCreds.AccessKeyID
	eSecKey := user.AwsCreds.SecretAccessKey

	accesskey,err := lib.DecryptSecret(eAccKey)
	if err != nil {
		http.Error(w, "Error decrypting secret: "+err.Error(), http.StatusInternalServerError)
		return
	}
	secretkey, err := lib.DecryptSecret(eSecKey)
	if err != nil {
		http.Error(w, "Error decrypting secret: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for _, service := range nvmsConfig.Services {
		/* So here we need to deploy 2x services into a vm, for now just a basic ec2, that will then be run with X command and have the following ports opened publicly for it. */

		DeployAWSService(accesskey,secretkey,service, codebase)
		// TODO: Deploy MicroVM

		// Load Files and Extract

		// Execute
	}

	//TODO: Begin Generating a Resource Plan -> Send to Builder Route
	//TODO: Build VPC/Network, Configure Security Groups, Setup Load Balancers, Go down the line of the Resource Plan/NVMS Object
	//TODO: Validate Resources and Send Status -> Deployment Module
	//TODO: Config/Deploy MicroVM(FireCracker), Config Services, Setup Monitoring call portfolio route (Repository, ReadMe, NVMS)
	//TODO: Analyze Project (Get Details for Prompting, Read Playground Type from NVMS), Pull Templates from Portfolio, Pick appropriate template given args and build and send back.
	//TODO: Open appropriate connections for playground and rpovide to route, deployed.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Deploying Project"))
}

func DeployAWSService(accessKey, secretKey string, Service lib.Service,zipBall []byte) error {
	fileMap, serviceFolder, err := getServiceFolder(Service.Name,zipBall)
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Region:      aws.String("us-east-1"),
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	svc := ec2.New(sess)
	result, err := DeployEC2Instance(svc, Service.Name, "BytePort(NVMS)")
    if err != nil {
        fmt.Println("Got an error creating an instance: " +Service.Name)
        return err
    }

    fmt.Println("Created tagged instance with ID " + *result.Instances[0].InstanceId)
	return nil
}
func DeployEC2Instance(svc *ec2.EC2, name string, tag string) (*ec2.Reservation, error) {
	
	serviceBuildConfig :=  `#!/bin/bash
    aws s3 cp s3://%s/!!frontend.zip!! /home/ec2-user/
    cd /home/ec2-user
    unzip !!frontend.zip!!
    cd frontend
    npm install
    npm run build
    `
	result, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId: aws.String("ami-0f4e0a7f"),
		InstanceType: aws.String("t2.micro"),
		MaxCount: aws.Int64(1),
		MinCount: aws.Int64(1),
		UserData:     aws.String(base64.StdEncoding.EncodeToString([]byte(serviceBuildConfig))),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key: aws.String("Name"),
						Value: aws.String(tag),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
