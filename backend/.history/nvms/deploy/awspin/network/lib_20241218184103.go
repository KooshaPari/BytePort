package network

import aws "nvms/deploy/awspin"
type Client struct {
    config       aws.Config
    endpointURL string
}