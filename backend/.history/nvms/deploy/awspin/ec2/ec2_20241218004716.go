package ec2

import (
	"bytes"
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

func (c *Client) newRequest(ctx context.Context, method string, params map[string]string, body []byte) (*http.Request, error) {
    u, err := url.Parse(c.endpointURL)
    if err != nil {
        return nil, err
    }

    // Add EC2 API version
    params["Version"] = "2016-11-15"
    
    query := u.Query()
    for k, v := range params {
        query.Set(k, v)
    }
    u.RawQuery = query.Encode()
	fmt.Println("Request URL: ", u.String())
    req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    var awsDate aws.AwsDate
    awsDate.Time = time.Now()

    payloadHash := aws.GetPayloadHash(body)
    req.Header.Set("host", u.Host)
    req.Header.Set("content-length", fmt.Sprintf("%d", len(body)))
    req.Header.Set("x-amz-content-sha256", payloadHash)
    req.Header.Set("x-amz-date", awsDate.GetTime())
    req.Header.Set("x-amz-security-token", c.config.SessionToken)
    req.Header.Set("user-agent", "byteport")
    req.Header.Set("authorization", aws.GetAuthorizationHeader(&c.config, req, &awsDate, payloadHash))

    return req, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
    resp, err := spinhttp.Send(req)
    if err != nil {
		fmt.Println("Error sending request: ", err)
        return nil, fmt.Errorf("failed to send request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
		fmt.Println("Code: ", resp.StatusCode)
		fmt.Print("Response: ", resp)
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