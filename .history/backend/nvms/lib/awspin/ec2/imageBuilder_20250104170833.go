package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
	"nvms/models"
	"time"
)
 
func (c *Client) CreatePipeline(ctx context.Context, name string, infraArn string) (*CreatePipelineResponse, error) {
	params := map[string]string{
		"clientToken": time.Now().String(),
	 	"infrastructureConfigurationArn": infraArn,
		"name": name,
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
func (c *Client) ExecuteImgPipeline(ctx context.Context, pipelineArn string) (*ExecuteImgPipelineResponse, error) {
	params := map[string]string{
		"clientToken": time.Now().String(),
		"imagePipelineArn": pipelineArn,
	}
	req, err := c.newRequest(ctx, "POST", params, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ExecuteImgPipelineResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
 
func (c *Client) CreateInfrastructureConfiguration(ctx context.Context, name string, instProfName string) (*CreateInfrastructureConfigurationResponse, error){
	// append /CreateInfrastructureConfiguration to the URL

	c.endpointURL += "/CreateInfrastructureConfiguration"
	params := map[string]string{
		"clientToken": time.Now().String(),
		"instanceProfileName": instProfName,
		"name": name,
	}
	req, err := c.newRequest(ctx, "PUT", params, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	// revert the URL back to the original
	c.endpointURL = c.endpointURL[:len(c.endpointURL)-30]
	defer resp.Body.Close()

	var result CreateInfrastructureConfigurationResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
 
func (c *Client) DeleteInfrastructureConfiguration(ctx context.Context, name string, infraARN string) (*DeleteInfrastructureConfigurationResponse, error){
	c.endpointURL += "/DeleteInfrastructureConfiguration"
	params := map[string]string{
		"clientToken": time.Now().String(),
		"infrastructureConfigArn": infraARN,
 
	}
	req, err := c.newRequest(ctx, "DELETE", params, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	c.endpointURL = c.endpointURL[:len(c.endpointURL)-30]
	defer resp.Body.Close()

	var result DeleteInfrastructureConfigurationResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
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
*/
func (c *Client) CreateImageRecipe(ctx context.Context, components []models.ImageComponent,baseArn string, name string) (*CreateImageRecipeResponse, error){
	body := CreateImageRecipeRequest{
		Components: components,
		ClientToken: time.Now().String(),
		SemanticVersion: "1.0.0",
		Name: name,
		ParentImage: baseArn,}
	req, err := c.newRequest(ctx, "PUT", nil, body)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	

}
//DELETE /DeleteImageRecipe?imageRecipeArn=imageRecipeArn HTTP/1.1
func (c *Client) DeleteImageRecipe(ctx context.Context) (*CreateImageRecipeResponse, error){}