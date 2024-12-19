package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	aws "nvms/deploy/awspin"
	"time"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)

// NewEC2 creates a new EC2 Client
func NewEC2(config aws.Config) (*Client, error) {
    u, err := url.Parse(config.Endpoint)
    if err != nil {
        return nil, fmt.Errorf("failed to parse endpoint: %w", err)
    }

    client := &Client{
        config:      config,
        endpointURL: u.String(),
    }

    return client, nil
}

//

//

func (c *Client) do(req *http.Request) (*http.Response, error) {
    resp, err := spinhttp.Send(req)
    if err != nil {
		fmt.Println("Error sending request: ", err)
        return nil, fmt.Errorf("failed to send request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
		fmt.Println("Code: ", resp.StatusCode)
		fmt.Println("Response: ", resp)
        var errorResponse aws.ErrorResponse
        if err := xml.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			fmt.Println("Error parsing response: ", err)
            return nil, fmt.Errorf("failed to parse error response: %w", err)
        }
		fmt.Println("Error response: ", errorResponse)
        return nil, errorResponse
    }
	fmt.Println("Request sent successfully")
    return resp, nil
}

// RunInstances launches new EC2 instances
func (c *Client) RunInstances(ctx context.Context, params map[string]string) (*RunInstancesResponse, error) {
    params["Action"] = "RunInstances"
	fmt.Println("Creating instance: ", params)
    
    req, err := c.newRequest(ctx, "POST", params, nil)
    if err != nil {
		fmt.Println("Error creating request: ", err)
        return nil, err
    }

    resp, err := c.do(req)
    if err != nil {
		fmt.Println("Error creating instance: ", err)
        return nil, err
    }
    defer resp.Body.Close()

    var result RunInstancesResponse
    if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error decoding response: ", err)
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
	fmt.Println("Instance created: ", result)
    return &result, nil
}

// DescribeInstances gets information about EC2 instances
func (c *Client) DescribeInstances(ctx context.Context, instanceIds []string) (*DescribeInstancesResponse, error) {
    params := map[string]string{
        "Action": "DescribeInstances",
    }
    
    for i, id := range instanceIds {
        params[fmt.Sprintf("InstanceId.%d", i+1)] = id
    }

    req, err := c.newRequest(ctx, "GET", params, nil)
    if err != nil {
        return nil, err
    }

    resp, err := c.do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result DescribeInstancesResponse
    if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return &result, nil
}

// TerminateInstances terminates EC2 instances
func (c *Client) TerminateInstances(ctx context.Context, instanceIds []string) error {
    params := map[string]string{
        "Action": "TerminateInstances",
    }
    
    for i, id := range instanceIds {
        params[fmt.Sprintf("InstanceId.%d", i+1)] = id
    }

    req, err := c.newRequest(ctx, "POST", params, nil)
    if err != nil {
        return err
    }

    resp, err := c.do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}