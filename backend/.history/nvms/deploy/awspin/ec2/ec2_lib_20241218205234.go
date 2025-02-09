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
    InstanceId string `xml:"instanceId"`
    ImageId    string `xml:"imageId"`
    State      struct {
        Code int    `xml:"code"`
        Name string `xml:"name"`
    } `xml:"instanceState"`
    PrivateDnsName string `xml:"privateDnsName"`
    DnsName        string `xml:"dnsName"`
    Reason         string `xml:"reason"`
    KeyName        string `xml:"keyName"`
    AmiLaunchIndex int    `xml:"amiLaunchIndex"`
    ProductCodes   []string `xml:"productCodes"`
    InstanceType   string `xml:"instanceType"`
    LaunchTime     string `xml:"launchTime"`
    Placement      struct {
        AvailabilityZone string `xml:"availabilityZone"`
        GroupName        string `xml:"groupName"`
    } `xml:"placement"`
    Monitoring struct {
        State string `xml:"state"`
    } `xml:"monitoring"`
    SubnetId          string `xml:"subnetId"`
    VpcId             string `xml:"vpcId"`
    PrivateIpAddress  string `xml:"privateIpAddress"`
    SourceDestCheck   bool   `xml:"sourceDestCheck"`
    GroupSet          []struct {
        GroupId   string `xml:"groupId"`
        GroupName string `xml:"groupName"`
    } `xml:"groupSet>item"`
    Architecture      string `xml:"architecture"`
    RootDeviceType    string `xml:"rootDeviceType"`
    RootDeviceName    string `xml:"rootDeviceName"`
    
}
/*
<RunInstancesResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
  <requestId>59dbff89-35bd-4eac-99ed-be587EXAMPLE</requestId>
  <reservationId>r-1234567890abcdef0</reservationId>
  <ownerId>123456789012</ownerId>
  <groupSet>
    <item>
      <groupId>sg-1234567890abcdef0</groupId>
      <groupName>my-security-group</groupName>
    </item>
  </groupSet>
  <instancesSet>
    <item>
      <instanceId>i-1234567890abcdef0</instanceId>
      <imageId>ami-1234567890abcdef0</imageId>
      <instanceState>
        <code>0</code>
        <name>pending</name>
      </instanceState>
      <privateDnsName>ip-10-0-0-100.ec2.internal</privateDnsName>
      <dnsName/>
      <reason/>
      <keyName>my-key-pair</keyName>
      <amiLaunchIndex>0</amiLaunchIndex>
      <productCodes/>
      <instanceType>t2.micro</instanceType>
      <launchTime>2023-04-01T12:00:00.000Z</launchTime>
      <placement>
        <availabilityZone>us-east-1a</availabilityZone>
        <groupName/>
      </placement>
      <monitoring>
        <state>disabled</state>
      </monitoring>
      <subnetId>subnet-1234567890abcdef0</subnetId>
      <vpcId>vpc-1234567890abcdef0</vpcId>
      <privateIpAddress>10.0.0.100</privateIpAddress>
      <sourceDestCheck>true</sourceDestCheck>
      <groupSet>
        <item>
          <groupId>sg-1234567890abcdef0</groupId>
          <groupName>my-security-group</groupName>
        </item>
      </groupSet>
      <architecture>x86_64</architecture>
      <rootDeviceType>ebs</rootDeviceType>
      <rootDeviceName>/dev/xvda</rootDeviceName>
      <blockDeviceMapping>
        <item>
          <deviceName>/dev/xvda</deviceName>
          <ebs>
            <volumeId>vol-1234567890abcdef0</volumeId>
            <status>attaching</status>
            <attachTime>2023-04-01T12:00:00.000Z</attachTime>
            <deleteOnTermination>true</deleteOnTermination>
          </ebs>
        </item>
      </blockDeviceMapping>
      <virtualizationType>hvm</virtualizationType>
      <clientToken>abcde1234567890123</clientToken>
      <tagSet>
        <item>
          <key>Name</key>
          <value>MyInstance</value>
        </item>
      </tagSet>
      <hypervisor>xen</hypervisor>
      <networkInterfaceSet>
        <item>
          <networkInterfaceId>eni-1234567890abcdef0</networkInterfaceId>
          <subnetId>subnet-1234567890abcdef0</subnetId>
          <vpcId>vpc-1234567890abcdef0</vpcId>
          <description>Primary network interface</description>
          <ownerId>123456789012</ownerId>
          <status>in-use</status>
          <macAddress>0e:9d:c8:a5:42:e5</macAddress>
          <privateIpAddress>10.0.0.100</privateIpAddress>
          <sourceDestCheck>true</sourceDestCheck>
          <groupSet>
            <item>
              <groupId>sg-1234567890abcdef0</groupId>
              <groupName>my-security-group</groupName>
            </item>
          </groupSet>
          <attachment>
            <attachmentId>eni-attach-1234567890abcdef0</attachmentId>
            <deviceIndex>0</deviceIndex>
            <status>attaching</status>
            <attachTime>2023-04-01T12:00:00.000Z</attachTime>
            <deleteOnTermination>true</deleteOnTermination>
          </attachment>
          <privateIpAddressesSet>
            <item>
              <privateIpAddress>10.0.0.100</privateIpAddress>
              <primary>true</primary>
            </item>
          </privateIpAddressesSet>
        </item>
      </networkInterfaceSet>
      <ebsOptimized>false</ebsOptimized>
    </item>
  </instancesSet>
</RunInstancesResponse>
*/

// RunInstancesResponse represents the response from RunInstances
type RunInstancesResponse struct {
    XMLName      xml.Name `xml:"RunInstancesResponse"`
    ReservationId string   `xml:"reservationId"`
    OwnerId       string   `xml:"ownerId"`
    Instances     []Instance `xml:"instancesSet>item"`
    
}

// DescribeInstancesResponse represents the response from DescribeInstances
type DescribeInstancesResponse struct {
    XMLName     xml.Name   `xml:"DescribeInstancesResponse"`
    Reservations []struct {
        Instances []Instance `xml:"instancesSet>item"`
    } `xml:"reservationSet>item"`
}
