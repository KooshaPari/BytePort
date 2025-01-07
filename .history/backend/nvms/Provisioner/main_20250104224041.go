package main

import (
	"fmt"
	"net/http"
	"nvms/lib"
	"nvms/models"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		// Receive Proj-Obj, ZipBall, push to S3, provision general needed services/net return resource graph
		var project models.Project; var zipBall []byte; var deployID string
		accesskey,secretkey, err := lib.GetAWSCredentials(project.User) 
		if err != nil {
			http.Error(w, "Error getting AWS credentials", http.StatusInternalServerError)
			return
		}
		bucket, err := lib.PushToS3(zipBall,  accesskey, secretkey, project.Name)
		if err != nil {
		fmt.Println("Error pushing to S3: ", err)
		http.Error(w, "Error pushing to S3: "+err.Error(), http.StatusInternalServerError)
		return
		}
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
		// Create ECR Repo
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


		// Start Img Build



		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello Fermyon!")
	})
}

func main() {}
