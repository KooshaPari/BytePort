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
func (c *Client) newRequest(ctx context.Context, method string, params map[string]string, body []byte) (*http.Request, error) {
    u, err := url.Parse(c.endpointURL)
    if err != nil {
        return nil, err
    }

    var awsDate AwsDate
    awsDate.Time = time.Now()

    // Add required AWS Query API parameters
    params["Version"] = "2016-11-15"
    params["X-Amz-Algorithm"] = "AWS4-HMAC-SHA256"
    params["X-Amz-Date"] = awsDate.GetTime()
    
    // Build credential scope
    credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request",
        awsDate.GetDate(),
        c.config.Region,
        c.config.Service)
    
    params["X-Amz-Credential"] = fmt.Sprintf("%s/%s",
        c.config.AccessKeyId,
        credentialScope)

    // Set signed headers
    params["X-Amz-SignedHeaders"] = "host"

    // Add security token if present
    if c.config.SessionToken != "" {
        params["X-Amz-Security-Token"] = c.config.SessionToken
    }

    // Build canonical query string for signing
    canonicalQueryString := GetCanonicalQueryString(params)

    // Create string to sign
    canonicalRequest := strings.Join([]string{
        method,
        "/",
        canonicalQueryString,
        fmt.Sprintf("host:%s\n", u.Host),  // Canonical headers
        "host",                            // Signed headers
        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // Empty payload hash
    }, "\n")

    stringToSign := strings.Join([]string{
        "AWS4-HMAC-SHA256",
        awsDate.GetTime(),
        credentialScope,
        getSHA256Hash([]byte(canonicalRequest)),
    }, "\n")

    // Calculate signature
    dateKey := hmacSHA256([]byte("AWS4"+c.config.SecretAccessKey), []byte(awsDate.GetDate()))
    regionKey := hmacSHA256(dateKey, []byte(c.config.Region))
    serviceKey := hmacSHA256(regionKey, []byte(c.config.Service))
    signingKey := hmacSHA256(serviceKey, []byte("aws4_request"))
    signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

    // Add signature to parameters
    params["X-Amz-Signature"] = signature

    // Build final URL with all parameters
    query := u.Query()
    for k, v := range params {
        query.Set(k, v)
    }
    u.RawQuery = query.Encode()

    fmt.Printf("Request URL: %s\n", u.String())

    // Create request with minimal headers
    req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("host", u.Host)
    req.Header.Set("user-agent", "byteport")

    fmt.Printf("Request headers: %+v\n", req.Header)
    return req, nil
}

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