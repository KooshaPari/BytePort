package ec2

import (
	"encoding/xml"
	aws "nvms/deploy/awspin"
)

// Client provides an interface for interacting with the EC2 API
type Client struct {
    config       aws.Config
    endpointURL string
}


// RunInstancesResponse represents the response from RunInstances
type RunInstancesResponse struct {
    XMLName   xml.Name   `xml:"RunInstancesResponse"`
    Instances []Instance `xml:"instancesSet>item"`
}

// DescribeInstancesResponse represents the response from DescribeInstances
type DescribeInstancesResponse struct {
    XMLName     xml.Name   `xml:"DescribeInstancesResponse"`
    Reservations []struct {
        Instances []Instance `xml:"instancesSet>item"`
    } `xml:"reservationSet>item"`
}
