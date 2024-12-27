package projectManager 
spinhttp "github.com/fermyon/spin-go-sdk/http"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)




func TerminateProject(w http.ResponseWriter, r *http.Request) {
 /*Get Project, User from Req -> Deployments from DeploymentsJSON, loop thru call a terminate resource func(analyze service type choose appropriate termination function)*/
 var project models.Project; var user models.User;
 project, user, err := readBody(w, r)
 if err != nil {
	 return
 }
 var deployments map[string]Instance
 err = json.Unmarshal([]byte(project.DeploymentsJSON), &deployments)
 if err != nil {
	 http.Error(w, "Error parsing JSON", http.StatusBadRequest)
	 return
 }
 accessKey,secretKey, err := getAWSCredentials(user) 
 for _, deployment := range deployments {
	 terminateResource(deployment)
 }

}
func termi

}