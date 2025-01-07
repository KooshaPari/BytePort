package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
)
/* Request Example
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
}*/
func (c *Client) CreatePipeline(ctx context.Context, params map[string]string) (*CreatePipelineResponse, error) {
	
	
}