package ec2
/*
HTTP/1.1 200 OK
Server: Server
Date: Fri, 16 Dec 2016 20:12:56 GMT
Content-Type: application/x-amz-json-1.1
Content-Length: 786
Connection: keep-alive
x-amzn-RequestId: 084038f1-c3cc-11e6-8d10-9da51cf53fd3

{
  "image": {
    "imageId": {
      "imageDigest": "sha256:f1d4ae3f7261a72e98c6ebefe9985cf10a0ea5bd762585a43e0700ed99863807",
      "imageTag": "2016.09"
    },
    "imageManifest": "{\n   \"schemaVersion\": 2,\n   \"mediaType\": \"application/vnd.docker.distribution.manifest.v2+json\",\n   \"config\": {\n      \"mediaType\": \"application/vnd.docker.container.image.v1+json\",\n      \"size\": 1486,\n      \"digest\": \"sha256:5b52b314511a611975c2c65e695d920acdf8ae8848fe0ef00b7d018d1f118b64\"\n   },\n   \"layers\": [\n      {\n         \"mediaType\": \"application/vnd.docker.image.rootfs.diff.tar.gzip\",\n         \"size\": 91768077,\n         \"digest\": \"sha256:8e3fa21c4cc40232e835a6761332d225c7af3235c5755f44ada2ed9d0e4ab7e8\"\n      }\n   ]\n}\n",
    "registryId": "012345678910",
    "repositoryName": "amazonlinux"
  }
}
*/
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


/*
HTTP/1.1 200
Content-type: application/json

{
   "clientToken": "string",
   "containerRecipeArn": "string",
   "requestId": "string"
}*/
type CreateContainerRecipeResponse struct {
	ClientToken string `json:"clientToken"`
	ContainerRecipeArn string `json:"containerRecipeArn"`
	RequestId string `json:"requestId"`
}
/*
{
   "containerRecipeArn": "string",
   "requestId": "string"
}*/
type DeleteContainerRecipeResponse struct {
	ContainerRecipeArn string `json:"containerRecipeArn"`

}