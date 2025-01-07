package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
)

func (c *Client) PutImage(ctx context.Context, params map[string]string) (*PutImageResponse, error) {
	params["Action"] = "PutImage"
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
func (c *Client) CreateRepository(ctx context.Context, params map[string]string) (*CreateRepositoryResponse, error) {
	params["Action"] = "CreateRepository"
	req, err := c.newRequest(ctx, "POST", params, nil)
	if err != nil {
		return nil, err
	}

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
func (c *Client) DeleteRepository(ctx context.Context, params map[string]string) (*DeleteRepositoryResponse, error) {
	params["Action"] = "DeleteRepository"
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