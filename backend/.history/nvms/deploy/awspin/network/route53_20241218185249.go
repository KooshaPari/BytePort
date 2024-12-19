package network

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
func NewRoute53(config aws.Config) (*Client, error) {
	// Create a new  Route53 client
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

func createHostedZone() {
	payload := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<CreateHostedZoneRequest xmlns="https://route53.amazonaws.com/doc/2013-04-01/">
  <Name>%s</Name>
  <CallerReference>%d</CallerReference>
  <HostedZoneConfig>
    <PrivateZone>true</PrivateZone>
  </HostedZoneConfig>
  <VPC>
    <VPCRegion>%s</VPCRegion>
    <VPCId>%s</VPCId>
  </VPC>
</CreateHostedZoneRequest>`, domainName, time.Now().Unix(), region, vpcId)
	// Create a new hosted zone
}
func createRecordSet() {
	// Create a new record set
}
func updateRecordSet() {
	// Update a record set
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

func (c *Client) newRequest(ctx context.Context, method, path string, body []byte) (*http.Request, error) {


	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	

	var awsDate  aws.AwsDate
	awsDate.Time = time.Now()

	payloadHash := aws.GetPayloadHash(body)
	req.Header.Set("host", u.Host)
	req.Header.Set("content-length", fmt.Sprintf("%d", len(body)))
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("x-amz-date", awsDate.GetTime())
	req.Header.Set("x-amz-security-token", c.config.SessionToken)
	req.Header.Set("user-agent", "byteport")
	req.Header.Set("authorization",  aws.GetAuthorizationHeader(&c.config, req, &awsDate, payloadHash))
	fmt.Println("Request: ", req)
	return req, nil
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