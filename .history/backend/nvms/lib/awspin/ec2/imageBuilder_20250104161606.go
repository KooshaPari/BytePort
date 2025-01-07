package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
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
	defer resp.Body.Close()

	var result CreateInfrastructureConfigurationResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
func (c *Client) DeleteInfrastructureConfiguration(ctx context.Context, name string, instProfName string) (*DeleteInfrastructureConfigurationResponse, error){
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
	defer resp.Body.Close()

	var result CreateInfrastructureConfigurationResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}