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
		/*ecrRepo, err := lib.CreateECRRepo(project.Name, accesskey, secretkey)
		if err != nil {
			fmt.Println("Error creating ECR Repo: ", err)
			http.Error(w, "Error creating ECR Repo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Build Image Builder Pipeline
		servImgPipeline, err := lib.CreateImgPipeline(project.Name, accesskey, secretkey)
		if err != nil {
			fmt.Println("Error creating Image Builder Pipeline: ", err)
			http.Error(w, "Error creating Image Builder Pipeline: "+err.Error(), http.StatusInternalServerError)
			return
		}*/
		

		// Start Img Build



		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello Fermyon!")
	})
}

func main() {}
