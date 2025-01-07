package ecr

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
)
type PutRequest struct {
   RepositoryName string `json:"repositoryName"`
   ImageManifest string `json:"imageManifest"`
}
func (c *Client) PutImage(ctx context.Context, name string, manifest string) (*PutImageResponse, error) {
	reqBody := PutRequest{
		RepositoryName: name,
		ImageManifest: manifest,
	}
   body, err := json.Marshal(reqBody)
   if err != nil {
      return nil, err
   }
	req, err := c.newRequest(ctx, "POST", nil, body, "AmazonEC2ContainerRegistry_V20150921.PutImage")
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
	requestBody := struct {
        RepositoryName string `json:"repositoryName"`
    }{
        RepositoryName: name,
    }
    
    bodyJSON, err := json.Marshal(requestBody)
    if err != nil {
        return nil, err
    }
    req, err := c.newRequest(ctx, "POST", nil, bodyJSON, "AmazonEC2ContainerRegistry_V20150921.CreateRepository")
    if err != nil {
        return nil, err
    }
	fmt.Println("Request created: ", req)
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
 
	var result CreateRepositoryResponse
	if err := json.Unmarshal(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
func (c *Client) DeleteRepository(ctx context.Context, name string) (*DeleteRepositoryResponse, error) {
	reqBody := struct{
      RepositoryName string `json:"repositoryName"`
      Force string `json:"Force"`
   }{ 
 
		RepositoryName: name,
		Force: "true",
	}
   bodyJSON, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }
	req, err := c.newRequest(ctx, "POST", nil, bodyJSON, "AmazonEC2ContainerRegistry_V20150921.DeleteRepository")
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
 