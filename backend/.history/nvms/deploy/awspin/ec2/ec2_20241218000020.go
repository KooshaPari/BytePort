package ec2


import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	aws "nvms/deploy/awspin"
	"time"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)

// Client provides an interface for interacting with the S3 API.
type Client struct {
	config       aws.Config
	endpointURL string
}

// New creates a new Client.
func NewS3(config  aws.Config) (*Client, error) {
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

 
// buildEndpoint returns an endpoint
func (c *Client) buildEndpoint(bucketName, path string) (string, error) {
	u, err := url.Parse(c.endpointURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse endpoint: %w", err)
	}
	if bucketName != "" {
		u.Host = bucketName + "." + u.Host
	}
	return u.JoinPath(path).String(), nil
}

func (c *Client) newRequest(ctx context.Context, method, bucketName, path string, body []byte) (*http.Request, error) {
	endpointURL, err := c.buildEndpoint(bucketName, path)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(endpointURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var awsDate  aws.AwsDate
	awsDate.Time = time.Now()

	// Set the AWS authentication headers
	payloadHash := aws.GetPayloadHash(body)
	req.Header.Set("host", u.Host)
	req.Header.Set("content-length", fmt.Sprintf("%d", len(body)))
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("x-amz-date", awsDate.GetTime())
	req.Header.Set("x-amz-security-token", c.config.SessionToken)
	// Optional
	req.Header.Set("user-agent", "spin-s3")
	req.Header.Set("authorization",  aws.GetAuthorizationHeader(&c.config, req, &awsDate, payloadHash))

	return req, nil
}
func (c *Client) CreateBucket(ctx context.Context, name string) error {
    fmt.Println("Creating bucket: ", name)
    req, err := c.newRequest(ctx, http.MethodPut, "", name, nil)
    if err != nil {
        fmt.Println("Error creating request: ", err)
        return err
    }

    resp, err := c.do(req)
    if err != nil {
        fmt.Println("Error creating bucket: ", err)
        return err
    }
    defer resp.Body.Close()
    return nil
}

// do sends the request and handles any error response.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := spinhttp.Send(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Only checking for a status of 200 feels too specific.
	if resp.StatusCode != http.StatusOK {
		var errorResponse  aws.ErrorResponse
		if err := xml.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return nil, errorResponse
	}
	return resp, nil
}
