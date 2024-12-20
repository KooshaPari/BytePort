package ec2

import (
	"encoding/xml"
	aws "nvms/lib/awspin"
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
type DescribeVpcsResponse struct {
    XMLName     xml.Name   `xml:"DescribeVpcsResponse"`
    Vpcs []struct {
        VpcId string `xml:"vpcId"`
        OwnerId string `xml:"ownerId"`
        State string `xml:"state"`
        CidrBlock string `xml:"cidrBlock"`
        CidrBlockAssociationSet []struct {
            CidrBlock string `xml:"cidrBlock"`
            AssociationId string `xml:"associationId"`
            CidrBlockState struct {
                State string `xml:"state"`
            } `xml:"cidrBlockState"`
        } `xml:"cidrBlockAssociationSet>item"`
        DhcpOptionsId string `xml:"dhcpOptionsId"`
        TagSet []struct {
            Key string `xml:"key"`
            Value string `xml:"value"`
        } `xml:"tagSet>item"`
        InstanceTenancy string `xml:"instanceTenancy"`
        IsDefault bool `xml:"isDefault"`
    } `xml:"vpcSet>item"`
}
 
type DescribeSecurityGroupsResponse struct {
    XMLName xml.Name `xml:"DescribeSecurityGroupsResponse"`
    SecurityGroupInfo struct {
        Item struct {
            OwnerId string `xml:"ownerId"`
            GroupId string `xml:"groupId"`
            GroupName string `xml:"groupName"`
            GroupDescription string `xml:"groupDescription"`
            VpcId string `xml:"vpcId"`
            IpPermissions []struct {
                Item struct {
                    IpProtocol string `xml:"ipProtocol"`
                    FromPort int `xml:"fromPort"`
                    ToPort int `xml:"toPort"`
                    Groups []struct {
                        Item struct {
                            SecurityGroupRuleId string `xml:"securityGroupRuleId"`
                            UserId string `xml:"userId"`
                            GroupId string `xml:"groupId"`
                            VpcId string `xml:"vpcId"`
                            VpcPeeringConnectionId string `xml:"vpcPeeringConnectionId"`
                            PeeringStatus string `xml:"peeringStatus"`
                        } `xml:"item"`
                    } `xml:"groups"`
                    IpRanges []struct {
                        Item struct {
                            CidrIp string `xml:"cidrIp"`
                        } `xml:"item"`
                    } `xml:"ipRanges"`
                    PrefixListIds []struct {
                        Item struct {
                            PrefixListId string `xml:"prefixListId"`
                        } `xml:"item"`
                    } `xml:"prefixListIds"`
                } `xml:"item"`
            } `xml:"ipPermissions"`
            IpPermissionsEgress []struct {
                Item struct {
                    IpProtocol string `xml:"ipProtocol"`
                    Groups []struct {
                        Item struct {
                            SecurityGroupRuleId string `xml:"securityGroupRuleId"`
                            UserId string `xml:"userId"`
                            GroupId string `xml:"groupId"`
                            VpcId string `xml:"vpcId"`
                            VpcPeeringConnectionId string `xml:"vpcPeeringConnectionId"`
                            PeeringStatus string `xml:"peeringStatus"`
                        } `xml:"item"`
                    } `xml:"groups"`
                    IpRanges []struct {
                        Item struct {
                            CidrIp string `xml:"cidrIp"`
                        } `xml:"item"`
                    } `xml:"ipRanges"`
                    PrefixListIds []struct {
                        Item struct {
                            PrefixListId string `xml:"prefixListId"`
                        } `xml:"item"`
                    } `xml:"prefixListIds"`
                } `xml:"item"`
            } `xml:"ipPermissionsEgress"`
        } `xml:"item"`
    } `xml:"securityGroupInfo"`
    

} 
 type DescribeSubnetsResponse struct {
    XMLName   xml.Name `xml:"DescribeSubnetsResponse"`
    SubnetSet []Subnet `xml:"subnetSet>item"`
}

type Subnet struct {
    SubnetId                     string `xml:"subnetId"`
    SubnetArn                    string `xml:"subnetArn"`
    State                        string `xml:"state"`
    OwnerId                      string `xml:"ownerId"`
    VpcId                        string `xml:"vpcId"`
    CidrBlock                    string `xml:"cidrBlock"`
    Ipv6CidrBlockAssociationSet  []Ipv6CidrBlockAssociation `xml:"ipv6CidrBlockAssociationSet>item"`
    AvailableIpAddressCount      int    `xml:"availableIpAddressCount"`
    AvailabilityZone             string `xml:"availabilityZone"`
    AvailabilityZoneId           string `xml:"availabilityZoneId"`
    DefaultForAz                 bool   `xml:"defaultForAz"`
    MapPublicIpOnLaunch          bool   `xml:"mapPublicIpOnLaunch"`
    AssignIpv6AddressOnCreation  bool   `xml:"assignIpv6AddressOnCreation"`
}

type Ipv6CidrBlockAssociation struct {
    Ipv6CidrBlock       string `xml:"ipv6CidrBlock"`
    AssociationId       string `xml:"associationId"`
    Ipv6CidrBlockState  struct {
        State string `xml:"state"`
    } `xml:"ipv6CidrBlockState"`
}