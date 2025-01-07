package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
)

 
func (c *Client) PutImage(ctx context.Context, name string, manifest string) (*PutImageResponse, error) {
	params := map[string]string{
		"X-Amz-Target": "AmazonEC2ContainerRegistry_V20150921.PutImage",
		"repositoryName": name,
		"imageManifest": manifest,
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

	var result PutImageResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
 /*POST / HTTP/1.1
Host: api.ecr.us-west-2.amazonaws.com
Accept-Encoding: identity
X-Amz-Target: AmazonEC2ContainerRegistry_V20150921.CreateRepository
Content-Type: application/x-amz-json-1.1
User-Agent: aws-cli/1.16.190 Python/3.6.1 Darwin/16.7.0 botocore/1.12.180
X-Amz-Date: 20190715T204735Z
Authorization: AUTHPARAMS
Content-Length: 33

{
   "repositoryName": "sample-repo"
}*/
func (c *Client) CreateRepository(ctx context.Context, name string) (*CreateRepositoryResponse, error) {
	params := map[string]string{
		"X-Amz-Target": "AmazonEC2ContainerRegistry_V20150921.CreateRepository",
		"repositoryName": name,
	}
	req, err := c.newRequest(ctx, "POST", params, nil)
	
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreateRepositoryResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
func (c *Client) DeleteRepository(ctx context.Context, name string) (*DeleteRepositoryResponse, error) {
	params := map[string]string{
		"X-Amz-Target": "AmazonEC2ContainerRegistry_V20150921.CreateRepository",
		"repositoryName": name,
		"Force": "true",
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

	var result DeleteRepositoryResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
/* PUT /CreateContainerRecipe HTTP/1.1
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
/*
func (c *Client) CreateContainerRecipe(ctx context.Context) (*CreateContainerRecipeResponse, error){}
// DELETE /DeleteContainerRecipe?containerRecipeArn=containerRecipeArn HTTP/1.1
func (c *Client) DeleteContainerRecipe(ctx context.Context) (*DeleteContainerRecipeResponse, error){}*/