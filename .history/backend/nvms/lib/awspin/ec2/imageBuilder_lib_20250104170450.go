package ec2

import "nvms/models"
 
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



 
type CreateInfrastructureConfigurationResponse struct{
   ClientToken string `json:"clientToken"`
   InfrastructureConfigurationArn string `json:"infrastructureConfigurationArn"`
   RequestId string `json:"requestId"`

}
 
type DeleteInfrastructureConfigurationResponse struct{
   InfrastructureConfigurationArn string `json:"infrastructureConfigurationArn"`
   RequestId string `json:"requestId"`
}
type ImageRecipe struct {
    ParentImage     string           // Base image
    Components      []models.ImageComponent // From buildpack
    WorkingDir     string           // Service path
    Tags           map[string]string
}
type CreateImageRecipeRequest struct {
   
type CreateImageRecipeResponse struct {
   ClientToken string `json:"clientToken"`
   ImageRecipeArn string `json:"imageRecipeArn"`
   RequestId string `json:"requestId"`

}
 
type DeleteImageRecipeResponse struct {
   ImageRecipeArn string `json:"imageRecipeArn"`
   RequestId string `json:"requestId"`
}
//type CreateImageComponentResponse struct {}
//type DeleteImageComponentResponse struct {}
