package main
import ("nvms/models"
"net/http"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http" )
func DeployProject(w http.ResponseWriter, r *http.Request) {
	var repository models.Repository;
	var project models.Project;
	ctx:= context.WithValue(r.Context(), "project", project)
	ctx = context.WithValue(ctx, "repository", repository)
}