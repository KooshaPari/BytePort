package vpc

import (
	aws "nvms/deploy/awspin"
)

// Client provides an interface for interacting with the EC2 API
type Client struct {
    config       aws.Config
    endpointURL string
}


type CreateVpcResponse struct {
    VpcId string `xml:"vpc>vpcId"`
    OwnerId string `xml:"vpc>ownerId"`
    state string `xml:"vpc>state"`
    