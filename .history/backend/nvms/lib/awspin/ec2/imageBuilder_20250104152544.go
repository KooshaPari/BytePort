package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
)

func (c *Client) CreatePipeline(ctx context.Context, params map[string]string) (*CreatePipelineResponse, error) {
	params["Action"] = "CreateImagePipeline"
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
func(c *Client) StartExecutePipeline(ctx context.Context, params map[string]string) (*StartPipelineExecResponse,error){
	params["Action"] = "CreateImagePipeline"
	req, err := c.newRequest(ctx, "PUT", params, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result StartPipelineExecResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
	
}
 