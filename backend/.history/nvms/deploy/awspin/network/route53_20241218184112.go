package network

import (
	"fmt"
	"net/url"
	aws "nvms/deploy/awspin"
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
	// Create a new hosted zone
}
func createRecordSet() {
	// Create a new record set
}
func updateRecordSet() {
	// Update a record set
}