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
func (c *Client) CreateImageRecipe(ctx context.Context) (*CreateImageRecipeResponse, error){}
func (c *Client) CreateImageRecipe(ctx context.Context) (*CreateImageRecipeResponse, error){}