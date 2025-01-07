package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"
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
func (c *Client) CreatePipeline(ctx context.Context, name string, repoName string, infraArn string) (*CreatePipelineResponse, error) {
	params := map[string]string{
		"clientToken": time.Now().String(),
	 	"infrastructureConfigurationArn": infraArn,
	}
		 // time.now 
	req, err := c.newRequest(ctx, "PUT", params, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreatePipelineResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
	
}