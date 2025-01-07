package ec2

import "nvms/models"
 
type CreatePipelineRequest struct {
   ClientToken string `json:"clientToken"`
   ContainerRecipeArn string `json:"containerRecipeArn"`
   Description string `json:"description"`
   DistributionConfigurationArn string `json:"distributionConfigurationArn"`
   EnhancedImageMetadataEnabled bool `json:"enhancedImageMetadataEnabled"`
   ExecutionRole string `json:"executionRole"`
   ImageRecipeArn string `json:"imageRecipeArn"`
   ImageScanningConfiguration struct {
      EcrConfiguration struct {
         ContainerTags []string `json:"containerTags"`
         RepositoryName string `json:"repositoryName"`
      } `json:"ecrConfiguration"`
      ImageScanningEnabled bool `json:"imageScanningEnabled"`
   } `json:"imageScanningConfiguration"`
   ImageTestsConfiguration struct {
      ImageTestsEnabled bool `json:"imageTestsEnabled"`
      TimeoutMinutes int `json:"timeoutMinutes"`
   } `json:"imageTestsConfiguration"`
   InfrastructureConfigurationArn string `json:"infrastructureConfigurationArn"`
   Name string `json:"name"`
   Schedule struct {
      PipelineExecutionStartCondition string `json:"pipelineExecutionStartCondition"`
      ScheduleExpression string `json:"scheduleExpression"`
      Timezone string `json:"timezone"`
   } `json:"schedule"`
   Status string `json:"status"`
   Tags map[string]string `json:"tags"`
   Workflows []struct {
      OnFailure string `json:"onFailure"`
      ParallelGroup string `json:"parallelGroup"`
      Parameters []struct {
         Name string `json:"name"`
         Value []string `json:"value"`
      } `json:"parameters"`
      WorkflowArn string `json:"workflowArn"`
   } `json:"workflows"`
}
/* 
PUT /CreateInfrastructureConfiguration HTTP/1.1
Content-type: application/json

{
   "clientToken": "string",
   "description": "string",
   "instanceMetadataOptions": { 
      "httpPutResponseHopLimit": number,
      "httpTokens": "string"
   },
   "instanceProfileName": "string",
   "instanceTypes": [ "string" ],
   "keyPair": "string",
   "logging": { 
      "s3Logs": { 
         "s3BucketName": "string",
         "s3KeyPrefix": "string"
      }
   },
   "name": "string",
   "placement": { 
      "availabilityZone": "string",
      "hostId": "string",
      "hostResourceGroupArn": "string",
      "tenancy": "string"
   },
   "resourceTags": { 
      "string" : "string" 
   },
   "securityGroupIds": [ "string" ],
   "snsTopicArn": "string",
   "subnetId": "string",
   "tags": { 
      "string" : "string" 
   },
   "terminateInstanceOnFailure": boolean
}*/

type CreateInfrastructureConfigurationRequest struct {
   ClientToken string `json:"clientToken"`
   Description string `json:"description"`
   InstanceMetadataOptions struct {
      HttpPutResponseHopLimit int `json:"httpPutResponseHopLimit"`
      HttpTokens string `json:"httpTokens"`
   } `json:"instanceMetadataOptions"`
   InstanceProfileName string `json:"instanceProfileName"`
   InstanceTypes []string `json:"instanceTypes"`
   KeyPair string `json:"keyPair"`
   Logging struct {
      S3Logs struct {
         S3BucketName string `json:"s3BucketName"`

         

}
type ExecuteImgPipelineRequest struct {}
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
 type ContainerRepo struct{
   RepositoryName string `json:"repositoryName"`
   Service string `json:"service"`
 }
type CreateContainerRecipeRequest struct {
   ClientToken string `json:"clientToken"`
   Components []ComponentsConfig `json:"components"`
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
   TargetRepository ContainerRepo  `json:"targetRepository"`
   WorkingDirectory string `json:"workingDirectory"`

}
type CreateContainerRecipeResponse struct {
   ClientToken string `json:"clientToken"`
   ContainerRecipeArn string `json:"containerRecipeArn"`
   RequestId string `json:"requestId"`
}
type ComponentsConfig struct {
   ComponentArn string `json:"componentArn"`
   Parameters []struct {
      Name string `json:"name"`
      Value []string `json:"value"`
   } `json:"parameters"`}
type DeleteContainerRecipeResponse struct {
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
