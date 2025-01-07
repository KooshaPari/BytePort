package ec2
 
type CreatePipelineResponse struct {
	ClientToken string `json:"clientToken"`
	ImagePipelineArn string `json:"imagePipelineArn"`
	RequestId string `json:"requestId"`
}
type DeletePipelineResponse struct {
   ImagePipelineArn string `json:"imagePipelineArn"`
   RequestId string `json:"requestId"`
}

type ExecuteImgPipelineResponse struct {
   ClientToken string   `json:"clientToken"`
   ImageBuildVersionArn string `json:"imageBuildVersionArn"`
   RequestId string `json:"requestId"`
}



/*{
   "clientToken": "string",
   "infrastructureConfigurationArn": "string",
   "requestId": "string"
}*/
type CreateInfrastructureConfigurationResponse struct{}
/*
HTTP/1.1 200
Content-type: application/json

{
   "infrastructureConfigurationArn": "string",
   "requestId": "string"
}*/
type DeleteInfrarstructureConfigurationResponse struct{
   InfrastructureConfigurationArn string `json:"infrastructureConfigurationArn"`
   RequestId string `json:"requestId"`
}