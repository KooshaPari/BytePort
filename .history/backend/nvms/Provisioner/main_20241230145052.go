package main

import (
	"fmt"
	"net/http"
	"nvms/models"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		// Receive Proj-Obj, ZipBall, push to S3, provision general needed services/net return resource graph
		var project models.Project; var zipBall []byte;
		bucket, err := lib.PushToS3(codebase, accesskey, secretkey, project.Name)
		if err != nil {
		fmt.Println("Error pushing to S3: ", err)
		http.Error(w, "Error pushing to S3: "+err.Error(), http.StatusInternalServerError)
		return
	}



		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello Fermyon!")
	})
}

func main() {}
