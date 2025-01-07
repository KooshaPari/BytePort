package ec2
 
type CreateRepositoryResponse struct {
	Repository struct {
		RepositoryArn string `json:"repositoryArn"`
		RegistryId string `json:"registryId"`
		RepositoryName string `json:"repositoryName"`
		RepositoryUri string `json:"repositoryUri"`
		CreatedAt float64 `json:"createdAt"`
		ImageTagMutability string `json:"imageTagMutability"`
		ImageScanningConfiguration struct {
			ScanOnPush bool `json:"scanOnPush"`
		} `json:"imageScanningConfiguration"`
	} `json:"repository"`
} 
 
type DeleteRepositoryResponse struct {
	Repository struct {
		RepositoryArn string `json:"repositoryArn"`
		RegistryId string `json:"registryId"`
		RepositoryName string `json:"repositoryName"`
		RepositoryUri string `json:"repositoryUri"`
		CreatedAt float64 `json:"createdAt"`
		ImageTagMutability string `json:"imageTagMutability"`
		ImageScanningConfiguration struct {
			ScanOnPush bool `json:"scanOnPush"`
		} `json:"imageScanningConfiguration"`
	} `json:"repository"`

}
/ 
type PutImageResponse struct {
	Image struct {
		ImageId struct {
			ImageDigest string `json:"imageDigest"`	
			ImageTag string `json:"imageTag"`
		} `json:"imageId"`
		ImageManifest string `json:"imageManifest"`
		RegistryId string `json:"registryId"`
		RepositoryName string `json:"repositoryName"`
	} `json:"image"`
}
