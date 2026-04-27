package route53

import aws "nvms/deploy/awspin"
type Client struct {
    config       aws.Config
    endpointURL string
}

  