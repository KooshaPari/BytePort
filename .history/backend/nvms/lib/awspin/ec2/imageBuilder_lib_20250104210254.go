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
 

type CreateImageRecipeResponse struct {
   ClientToken string `json:"clientToken"`
   ImageRecipeArn string `json:"imageRecipeArn"`
   RequestId string `json:"requestId"`

}
/*
PUT /CreateContainerRecipe HTTP/1.1
Content-type: application/json

{
   "clientToken": "string",
   "components": [ 
      { 
         "componentArn": "string",
         "parameters": [ 
            { 
               "name": "string",
               "value": [ "string" ]
            }
         ]
      }
   ],
   "containerType": "string",
   "description": "string",
   "dockerfileTemplateData": "string",
   "dockerfileTemplateUri": "string",
   "imageOsVersionOverride": "string",
   "instanceConfiguration": { 
      "blockDeviceMappings": [ 
         { 
            "deviceName": "string",
            "ebs": { 
               "deleteOnTermination": boolean,
               "encrypted": boolean,
               "iops": number,
               "kmsKeyId": "string",
               "snapshotId": "string",
               "throughput": number,
               "volumeSize": number,
               "volumeType": "string"
            },
            "noDevice": "string",
            "virtualName": "string"
         }
      ],
      "image": "string"
   },
   "kmsKeyId": "string",
   "name": "string",
   "parentImage": "string",
   "platformOverride": "string",
   "semanticVersion": "string",
   "tags": { 
      "string" : "string" 
   },
   "targetRepository": { 
      "repositoryName": "string",
      "service": "string"
   },
   "workingDirectory": "string"
}*/
type CreateContainerRecipeRequest struct {
   ClientToken string `json:"clientToken"`
   Components []models.ImageComponent `json:"components"`
   ContainerType string `json:"containerType"`
   Description string `json:"description"`
   DockerfileTemplateData string `json:"dockerfileTemplateData"`
   DockerfileTemplateUri string `json:"dockerfileTemplateUri"`
   ImageOsVersionOverride string `json:"imageOsVersionOverride"`
   InstanceConfiguration struct {
      BlockDeviceMappings []struct {
         DeviceName string `json:"deviceName"`
         Ebs struct {
            DeleteOnTermination bool `json:"deleteOnTermination"`
            Encrypted bool `json:"encrypted"`
            Iops int `json:"iops"`
            KmsKeyId string `json:"kmsKeyId"`
            SnapshotId string `json:"snapshotId"`
            Throughput int `json:"throughput"`
            VolumeSize int `json:"volumeSize"`
            VolumeType string `json:"volumeType"`
         } `json:"ebs"`
         NoDevice string `json:"noDevice"`
         VirtualName string `json:"virtualName"`
      } `json:"blockDeviceMappings"`
      Image string `json:"image"`
   } `json:"instanceConfiguration"`
   KmsKeyId string `json:"kmsKeyId"`
   Name string `json:"name"`
   ParentImage string `json:"parentImage"`
   PlatformOverride string `json:"platformOverride"`
   SemanticVersion string `json:"semanticVersion"`
   Tags map[string]string `json:"tags"`
   TargetRepository struct {
      RepositoryName string `json:"repositoryName"`
      Service string `json:"service"`
   } `json:"targetRepository"`
   WorkingDirectory string `json:"workingDirectory"`

}
type CreateContainerRecipeResponse struct {
   ClientToken string `json:"clientToken"`
   ContainerRecipeArn string `json:"containerRecipeArn"`
   RequestId string `json:"requestId"`
}

type CreateImageRecipeRequest struct {
   AdditionalInstanceConfiguration struct{
      SystemsManagerAgent struct{
         UninstallAfterBuild bool `json:"uninstallAfterBuild"`
      } `json:"systemsManagerAgent"`
      UserDataOverride string `json:"userDataOverride"`
      
   }
   BlockDeviceMappings struct{
      DeviceName string `json:"deviceName"`
      Ebs struct {
         DeleteOnTermination bool `json:"deleteOnTermination"`
         Encrypted bool `json:"encrypted"`
         Iops int `json:"iops"`
         KmsKeyId string `json:"kmsKeyId"`
         SnapshotId string `json:"snapshotId"`
         Throughput int `json:"throughput"`
         VolumeSize int `json:"volumeSize"`
         VolumeType string `json:"volumeType"`

      }
      NoDevice string `json:"noDevice"`
      VirtualName string `json:"virtualName"`
   } `json:"blockDeviceMappings"`
 
   ClientToken string `json:"clientToken"`
   Components []models.ImageComponent `json:"components"`
   Description string `json:"description"`
   Name string `json:"name"`
   ParentImage string `json:"parentImage"`
   SemanticVersion string `json:"semanticVersion"`
   Tags map[string]string `json:"tags"`
   WorkingDirectory string `json:"workingDirectory"`

}
type DeleteImageRecipeResponse struct {
   ImageRecipeArn string `json:"imageRecipeArn"`
   RequestId string `json:"requestId"`
}
 
type CreateImageComponentRequest struct {
   ChangeDescription string `json:"changeDescription"`
   ClientToken string `json:"clientToken"`
   Data string `json:"data"`
   Description string `json:"description"`
   KmsKeyId string `json:"kmsKeyId"`
   Name string `json:"name"`
   Platform string `json:"platform"`
   SemanticVersion string `json:"semanticVersion"`
   SupportedOsVersions []string `json:"supportedOsVersions"`
   Tags map[string]string `json:"tags"`
   Uri string `json:"uri"`

}
 
type CreateImageComponentResponse struct {
   ClientToken string `json:"clientToken"`
   ComponentBuildVersionArn string `json:"componentBuildVersionArn"`
   RequestId string `json:"requestId"`
}
//type DeleteImageComponentResponse struct {}
