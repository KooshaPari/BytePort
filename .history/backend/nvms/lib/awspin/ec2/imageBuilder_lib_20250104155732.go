package ec2
 
type CreatePipelineResponse struct {
	ClientToken string `json:"clientToken"`
	ImagePipelineArn string `json:"imagePipelineArn"`
	RequestId string `json:"requestId"`
}
type DeletePipelineResponse struct {
   imagePipelineArn string `json:"imagePipelineArn"`
   requestId string `json:"requestId"`
}

type Exec

/* Req Object
PUT /CreateImagePipeline HTTP/1.1
Content-type: application/json

{
   "clientToken": "string",
   "containerRecipeArn": "string",
   "description": "string",
   "distributionConfigurationArn": "string",
   "enhancedImageMetadataEnabled": boolean,
   "executionRole": "string",
   "imageRecipeArn": "string",
   "imageScanningConfiguration": { 
      "ecrConfiguration": { 
         "containerTags": [ "string" ],
         "repositoryName": "string"
      },
      "imageScanningEnabled": boolean
   },
   "imageTestsConfiguration": { 
      "imageTestsEnabled": boolean,
      "timeoutMinutes": number
   },
   "infrastructureConfigurationArn": "string",
   "name": "string",
   "schedule": { 
      "pipelineExecutionStartCondition": "string",
      "scheduleExpression": "string",
      "timezone": "string"
   },
   "status": "string",
   "tags": { 
      "string" : "string" 
   },
   "workflows": [ 
      { 
         "onFailure": "string",
         "parallelGroup": "string",
         "parameters": [ 
            { 
               "name": "string",
               "value": [ "string" ]
            }
         ],
         "workflowArn": "string"
      }
   ]
} */
