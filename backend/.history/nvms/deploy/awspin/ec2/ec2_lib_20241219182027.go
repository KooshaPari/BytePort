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
/*
<DescribeSecurityGroupsResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
    <requestId>edb7c570-be05-4192-bd1b-example</requestId>
    <securityGroupInfo>
        <item>
            <ownerId>123456789012</ownerId>
            <groupId>sg-1a2b3c4d</groupId>
            <groupName>MySecurityGroup</groupName>
            <groupDescription>MySecurityGroup</groupDescription>
            <vpcId>vpc-81326ae4</vpcId>
            <ipPermissions>
                <item>
                    <ipProtocol>tcp</ipProtocol>
                    <fromPort>22</fromPort>
                    <toPort>22</toPort>
                    <groups/>
                    <ipRanges>
                        <item>
                            <cidrIp>0.0.0.0/0</cidrIp>
                        </item>
                    </ipRanges>
                    <prefixListIds/>
                </item>
                <item>
                    <ipProtocol>icmp</ipProtocol>
                    <fromPort>-1</fromPort>
                    <toPort>-1</toPort>
                    <groups>
                        <item>
                            <securityGroupRuleId>sgr-abcdefghi01234560</securityGroupRuleId>
                            <userId>111222333444</userId>
                            <groupId>sg-11aa22bb</groupId>
                            <vpcId>vpc-dd326ab8</vpcId>
                            <vpcPeeringConnectionId>pcx-11223344</vpcPeeringConnectionId>
                            <peeringStatus>active</peeringStatus>
                        </item>
                    </groups>
                    <ipRanges/>
                    <prefixListIds/>
                </item>
            </ipPermissions>
            <ipPermissionsEgress>
                <item>
                    <ipProtocol>-1</ipProtocol>
                    <groups/>
                    <ipRanges>
                        <item>
                            <cidrIp>0.0.0.0/0</cidrIp>
                        </item>
                    </ipRanges>
                    <prefixListIds/>
                </item>
            </ipPermissionsEgress>
        </item>
    </securityGroupInfo>
</DescribeSecurityGroupsResponse>*/
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
/*<DescribeSubnetsResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
    <requestId>1927e20c-0ed0-4a02-a6d7-d955fbd2d13c</requestId>
    <subnetSet>
        <item>
            <subnetId>subnet-0bb1c79de301436ee</subnetId>
            <subnetArn>arn:aws:ec2:us-east-2:111122223333:subnet/subnet-0bb1c79de3EXAMPLE</subnetArn>
            <state>available</state>
            <ownerId>111122223333</ownerId>
            <vpcId>vpc-0ee975135dEXAMPLE</vpcId>
            <cidrBlock>10.0.2.0/24</cidrBlock>
            <ipv6CidrBlockAssociationSet/>
            <availableIpAddressCount>251</availableIpAddressCount>
            <availabilityZone>us-east-2c</availabilityZone>
            <availabilityZoneId>use2-az3</availabilityZoneId>
            <defaultForAz>false</defaultForAz>
            <mapPublicIpOnLaunch>false</mapPublicIpOnLaunch>
            <assignIpv6AddressOnCreation>false</assignIpv6AddressOnCreation>
        </item>
        <item>
            <subnetId>subnet-02bf4c428bf44ebce</subnetId>
            <subnetArn>arn:aws:ec2:us-east-2:111122223333:subnet/subnet-02bf4c428bEXAMPLE</subnetArn>
            <state>available</state>
            <ownerId>111122223333</ownerId>
            <vpcId>vpc-07e8ffd50fEXAMPLE</vpcId>
            <cidrBlock>10.0.0.0/24</cidrBlock>
            <ipv6CidrBlockAssociationSet>
                <item>
                    <ipv6CidrBlock>2600:1f16:115:200::/64</ipv6CidrBlock>
                    <associationId>subnet-cidr-assoc-002afb9f3cEXAMPLE</associationId>
                    <ipv6CidrBlockState>
                        <state>associated</state>
                    </ipv6CidrBlockState>
                </item>
            </ipv6CidrBlockAssociationSet>
            <availableIpAddressCount>251</availableIpAddressCount>
            <availabilityZone>us-east-2b</availabilityZone>
            <availabilityZoneId>use2-az2</availabilityZoneId>
            <defaultForAz>false</defaultForAz>
            <mapPublicIpOnLaunch>false</mapPublicIpOnLaunch>
            <assignIpv6AddressOnCreation>false</assignIpv6AddressOnCreation>
        </item>
    </subnetSet>
</DescribeSubnetsResponse>*/
type DescribeSubnetsResponse struct {
    XMLName xml.Name `xml:"DescribeSubnetsResponse"`
    SubnetSet []struct {
        Item struct {
            SubnetId string `xml:"subnetId"`
            SubnetArn string `xml:"subnetArn"`
            State string `xml:"state"`
            OwnerId string `xml:"ownerId"`
            VpcId string `xml:"vpcId"`
            CidrBlock string `xml:"cidrBlock"`
            Ipv6CidrBlockAssociationSet []struct {
                Item struct {
                    Ipv6CidrBlock string `xml:"ipv6CidrBlock"`
                    AssociationId string `xml:"associationId"`
                    Ipv6CidrBlockState struct {
                        State string `xml:"state"`
                    } `xml:"ipv6CidrBlockState"`
                } `xml:"item"`
            } `xml:"ipv6CidrBlockAssociationSet"`
            AvailableIpAddressCount int `xml:"availableIpAddressCount"`
            AvailabilityZone string `xml:"availabilityZone"`
            AvailabilityZoneId string `xml:"availabilityZoneId"`
            DefaultForAz bool `xml:"defaultForAz"`
            MapPublicIpOnLaunch bool `xml:"mapPublicIpOnLaunch"`
            AssignIpv6AddressOnCreation bool `xml:"assignIpv6AddressOnCreation"`
        } `xml:"item"`
    } `xml:"subnetSet"`
    

}