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

// Instance represents an EC2 instance
type Instance struct {
    InstanceId   string `xml:"instanceId"`
    InstanceType string `xml:"instanceType"`
    State        string `xml:"state>name"`
    PublicIP     string `xml:"publicIp"`
	PrivateIP	string `xml:"privateIp"`
	PublicDNS    string `xml:"publicDnsName"`
	PrivateDNS   string `xml:"privateDnsName"`
	KeyPairName  string `xml:"keyName"`
	SecurityGroups []string `xml:"groupSet>item>groupName"`
	SubnetId     string `xml:"subnetId"`
	R
    LaunchTime   string `xml:"launchTime"`
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
