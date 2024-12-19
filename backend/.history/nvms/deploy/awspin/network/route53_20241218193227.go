package network

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	aws "nvms/deploy/awspin"
	"strings"
	"time"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)

// NewRoute53 initializes a Route 53 client.
func NewRoute53(config aws.Config) (*Client, error) {
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

// CreateHostedZone creates a new private hosted zone.
func (c *Client)CreateHostedZone(ctx context.Context, domainName, region, vpcId string) (string,error) {
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

	resp, err := c.newRequest(ctx, http.MethodPost, "/2013-04-01/hostedzone", []byte(payload))
	if err != nil {
		return "",fmt.Errorf("failed to create hosted zone: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "",fmt.Errorf("failed to create hosted zone, status: %s", resp.Status)
	}


	type CreateHostedZoneResponse struct {
		HostedZone struct {
			ID string `xml:"Id"`
		} `xml:"HostedZone"`
	}
	var result CreateHostedZoneResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "",fmt.Errorf("failed to parse response: %w", err)
	}
	// /hostedzone/id -> id
	zoneID := strings.Split(result.HostedZone.ID, "/")[2]
	return zoneID,nil
}

// CreateRecordSet creates a new record set in the hosted zone.
func (c *Client) CreateRecordSet(ctx context.Context, hostedZoneID, name, recordType, value string, ttl int) error {
	payload := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ChangeResourceRecordSetsRequest xmlns="https://route53.amazonaws.com/doc/2013-04-01/">
  <ChangeBatch>
    <Changes>
      <Change>
        <Action>CREATE</Action>
        <ResourceRecordSet>
          <Name>%s</Name>
          <Type>%s</Type>
          <TTL>%d</TTL>
          <ResourceRecords>
            <ResourceRecord>
              <Value>%s</Value>
            </ResourceRecord>
          </ResourceRecords>
        </ResourceRecordSet>
      </Change>
    </Changes>
  </ChangeBatch>
</ChangeResourceRecordSetsRequest>`, name, recordType, ttl, value)

	path := fmt.Sprintf("/2013-04-01/hostedzone/%s/rrset", hostedZoneID)
	resp, err := c.newRequest(ctx, http.MethodPost, path, []byte(payload))
	if err != nil {
		return fmt.Errorf("failed to create record set: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create record set, status: %s", resp.Status)
	}

	fmt.Println("Record set created successfully.")
	return nil
}

// UpdateRecordSet updates an existing record set in the hosted zone.
func (c *Client) updateRecordSet(ctx context.Context, hostedZoneID, name, recordType, value string, ttl int) error {
	payload := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ChangeResourceRecordSetsRequest xmlns="https://route53.amazonaws.com/doc/2013-04-01/">
  <ChangeBatch>
    <Changes>
      <Change>
        <Action>UPSERT</Action>
        <ResourceRecordSet>
          <Name>%s</Name>
          <Type>%s</Type>
          <TTL>%d</TTL>
          <ResourceRecords>
            <ResourceRecord>
              <Value>%s</Value>
            </ResourceRecord>
          </ResourceRecords>
        </ResourceRecordSet>
      </Change>
    </Changes>
  </ChangeBatch>
</ChangeResourceRecordSetsRequest>`, name, recordType, ttl, value)

	path := fmt.Sprintf("/2013-04-01/hostedzone/%s/rrset", hostedZoneID)
	resp, err := c.newRequest(ctx, http.MethodPost, path, []byte(payload))
	if err != nil {
		return fmt.Errorf("failed to update record set: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update record set, status: %s", resp.Status)
	}

	fmt.Println("Record set updated successfully.")
	return nil
}

// buildEndpoint builds the request URL.
func (c *Client) buildEndpoint(path string) (string, error) {
	u, err := url.Parse(c.endpointURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse endpoint: %w", err)
	}
	return u.JoinPath(path).String(), nil
}

// newRequest builds and signs a new HTTP request.
func (c *Client) newRequest(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	endpoint, err := c.buildEndpoint(path)
	if err != nil {
		return nil, fmt.Errorf("failed to build endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	payloadHash := aws.GetPayloadHash(body)
	awsDate := aws.AwsDate{Time: time.Now()}
	req.Header.Set("host", req.URL.Host)
	req.Header.Set("content-length", fmt.Sprintf("%d", len(body)))
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("x-amz-date", awsDate.GetTime())
	if c.config.SessionToken != "" {
		req.Header.Set("x-amz-security-token", c.config.SessionToken)
	}
	req.Header.Set("authorization", aws.GetAuthorizationHeader(&c.config, req, &awsDate, payloadHash))

	return c.do(req)
}

// do sends the request and handles the response.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := spinhttp.Send(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResponse aws.ErrorResponse
		if err := xml.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return nil, fmt.Errorf("failed to parse error response: %w", err)
		}
		return nil, errorResponse
	}
	return resp, nil
}