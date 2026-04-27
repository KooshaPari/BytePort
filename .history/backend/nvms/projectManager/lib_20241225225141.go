package projectManager
type ProvisionerResponse struct {
	Nvms    string `json:"nvmsFile"`
	Readme  string `json:"Readme"`
	ZipBall []byte `json:"zipball"`
	FileMap []string `json:"fileMap"`
}
func readBody(w http.ResponseWriter, r *http.Request)(models.Project,modelss{ } 