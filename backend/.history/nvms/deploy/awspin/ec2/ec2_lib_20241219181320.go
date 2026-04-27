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
    BlockDeviceMapping []struct {
        DeviceName string `xml:"deviceName"`
        Ebs        struct {
            VolumeId            string `xml:"volumeId"`
            Status              string `xml:"status"`
            AttachTime          string `xml:"attachTime"`
            DeleteOnTermination bool   `xml:"deleteOnTermination"`
        } `xml:"ebs"`
    } `xml:"blockDeviceMapping>item"`
    VirtualizationType string `xml:"virtualizationType"`
    ClientToken        string `xml:"clientToken"`
    TagSet             []struct {
        Key   string `xml:"key"`
        Value string `xml:"value"`
    } `xml:"tagSet>item"`
    Hypervisor string `xml:"hypervisor"`
    NetworkInterfaceSet []struct {
        NetworkInterfaceId string `xml:"networkInterfaceId"`
        SubnetId           string `xml:"subnetId"`
        VpcId              string `xml:"vpcId"`
        Description        string `xml:"description"`
        OwnerId            string `xml:"ownerId"`
        Status             string `xml:"status"`
        MacAddress         string `xml:"macAddress"`
        PrivateIpAddress   string `xml:"privateIpAddress"`
        SourceDestCheck    bool   `xml:"sourceDestCheck"`
        GroupSet           []struct {
            GroupId   string `xml:"groupId"`
            GroupName string `xml:"groupName"`
        } `xml:"groupSet>item"`
        Attachment struct {
            AttachmentId         string `xml:"attachmentId"`
            DeviceIndex          int    `xml:"deviceIndex"`
            Status               string `xml:"status"`
            AttachTime           string `xml:"attachTime"`
            DeleteOnTermination  bool   `xml:"deleteOnTermination"`
        } `xml:"attachment"`
        PrivateIpAddressesSet []struct {
            PrivateIpAddress string `xml:"privateIpAddress"`
            Primary          bool   `xml:"primary"`
        } `xml:"privateIpAddressesSet>item"`
    } `xml:"networkInterfaceSet>item"`
    EbsOptimized bool `xml:"ebsOptimized"`

}
 
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
/*
<DescribeVpcsResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
    <requestId>8b67ac77-886c-4027-8f0e-d351f7fc9971</requestId>
    <vpcSet>
        <item>
            <vpcId>vpc-081ec835f3EXAMPLE</vpcId>
            <ownerId>111122223333</ownerId>
            <state>available</state>
            <cidrBlock>10.0.1.0/24</cidrBlock>
            <cidrBlockAssociationSet>
                <item>
                    <cidrBlock>10.0.1.0/24</cidrBlock>
                    <associationId>vpc-cidr-assoc-043f572c17EXAMPLE</associationId>
                    <cidrBlockState>
                        <state>associated</state>
                    </cidrBlockState>
                </item>
            </cidrBlockAssociationSet>
            <dhcpOptionsId>dopt-19edf471</dhcpOptionsId>
            <tagSet>
                <item>
                    <key>Name</key>
                    <value>MyVPC</value>
                </item>
            </tagSet>
            <instanceTenancy>default</instanceTenancy>
            <isDefault>false</isDefault>
        </item>
        <item>
            <vpcId>vpc-0ee975135dEXAMPLE</vpcId>
            <ownerId>111122223333</ownerId>
            <state>available</state>
            <cidrBlock>10.0.0.0/16</cidrBlock>
            <cidrBlockAssociationSet>
                <item>
                    <cidrBlock>10.0.0.0/16</cidrBlock>
                    <associationId>vpc-cidr-assoc-067c3a0a1fEXAMPLE</associationId>
                    <cidrBlockState>
                        <state>associated</state>
                    </cidrBlockState>
                </item>
            </cidrBlockAssociationSet>
            <dhcpOptionsId>dopt-f30d649a</dhcpOptionsId>
            <instanceTenancy>default</instanceTenancy>
            <isDefault>false</isDefault>
        </item>
        <item>
            <vpcId>vpc-06e4ab6c6cEXAMPLE</vpcId>
            <ownerId>123456789012</ownerId>
            <state>available</state>
            <cidrBlock>10.0.0.0/16</cidrBlock>
            <cidrBlockAssociationSet>
                <item>
                    <cidrBlock>10.0.0.0/16</cidrBlock>
                    <associationId>vpc-cidr-assoc-00b17b4eddEXAMPLE</associationId>
                    <cidrBlockState>
                        <state>associated</state>
                    </cidrBlockState>
                </item>
            </cidrBlockAssociationSet>
            <dhcpOptionsId>dopt-19edf471</dhcpOptionsId>
            <tagSet>
                <item>
                    <key>Name</key>
                    <value>Shared VPC</value>
                </item>
            </tagSet>
            <instanceTenancy>default</instanceTenancy>
            <isDefault>false</isDefault>
        </item>
    </vpcSet>
</DescribeVpcsResponse>
*/
type struct 