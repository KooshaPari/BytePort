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
    cidrBlock string `xml:"vpc>cidrBlock"`
    cidrBlockAssociationSet []struct {
        CidrBlock string `xml:"cidrBlock"`
        associationId string `xml:"associationId"`
        cidrBlockState struct {
            state string `xml:"state"`
            
    } `xml:"vpc>cidrBlockAssociationSet>item"`