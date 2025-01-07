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
 /*
PUT /CreateImageRecipe HTTP/1.1
Content-type: application/json

{
   "additionalInstanceConfiguration": { 
      "systemsManagerAgent": { 
         "uninstallAfterBuild": boolean
      },
      "userDataOverride": "string"
   },
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
   "description": "string",
   "name": "string",
   "parentImage": "string",
   "semanticVersion": "string",
   "tags": { 
      "string" : "string" 
   },
   "workingDirectory": "string"
}
   The request accepts the following data in JSON format.

additionalInstanceConfiguration
Specify additional settings and launch scripts for your build instances.

Type: AdditionalInstanceConfiguration object

Required: No

blockDeviceMappings
The block device mappings of the image recipe.

Type: Array of InstanceBlockDeviceMapping objects

Required: No

clientToken
Unique, case-sensitive identifier you provide to ensure idempotency of the request. For more information, see Ensuring idempotency in the Amazon EC2 API Reference.

Type: String

Length Constraints: Minimum length of 1. Maximum length of 36.

Required: Yes

components
The components included in the image recipe.

Type: Array of ComponentConfiguration objects

Array Members: Minimum number of 1 item.

Required: Yes

description
The description of the image recipe.

Type: String

Length Constraints: Minimum length of 1. Maximum length of 1024.

Required: No

name
The name of the image recipe.

Type: String

Pattern: ^[-_A-Za-z-0-9][-_A-Za-z0-9 ]{1,126}[-_A-Za-z-0-9]$

Required: Yes

parentImage
The base image of the image recipe. The value of the string can be the ARN of the base image or an AMI ID. The format for the ARN follows this example: arn:aws:imagebuilder:us-west-2:aws:image/windows-server-2016-english-full-base-x86/x.x.x. You can provide the specific version that you want to use, or you can use a wildcard in all of the fields. If you enter an AMI ID for the string value, you must have access to the AMI, and the AMI must be in the same Region in which you are using Image Builder.

Type: String

Length Constraints: Minimum length of 1. Maximum length of 1024.

Required: Yes

semanticVersion
The semantic version of the image recipe. This version follows the semantic version syntax.

Note
The semantic version has four nodes: <major>.<minor>.<patch>/<build>. You can assign values for the first three, and can filter on all of them.

Assignment: For the first three nodes you can assign any positive integer value, including zero, with an upper limit of 2^30-1, or 1073741823 for each node. Image Builder automatically assigns the build number to the fourth node.

Patterns: You can use any numeric pattern that adheres to the assignment requirements for the nodes that you can assign. For example, you might choose a software version pattern, such as 1.0.0, or a date, such as 2021.01.01.

Type: String

Pattern: ^[0-9]+\.[0-9]+\.[0-9]+$

Required: Yes

tags
The tags of the image recipe.

Type: String to string map

Map Entries: Maximum number of 50 items.

Key Length Constraints: Minimum length of 1. Maximum length of 128.

Key Pattern: ^(?!aws:)[a-zA-Z+-=._:/]+$

Value Length Constraints: Maximum length of 256.

Required: No

workingDirectory
The working directory used during build and test workflows.

Type: String

Length Constraints: Minimum length of 1. Maximum length of 1024.

Required: No
*/
type CreateImageRecipeRequest struct {
   AdditionalInstanceConfiguration AdditionalInstanceConfiguration `json:"additionalInstanceConfiguration"`
   BlockDeviceMappings []InstanceBlockDeviceMapping `json:"blockDeviceMappings"`
   
}
type DeleteImageRecipeResponse struct {
   ImageRecipeArn string `json:"imageRecipeArn"`
   RequestId string `json:"requestId"`
}
//type CreateImageComponentResponse struct {}
//type DeleteImageComponentResponse struct {}
